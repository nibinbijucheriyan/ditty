package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// RiskTicket represents a risk alert stored in the MCP server
type RiskTicket struct {
	ID         string    `json:"id"`
	File       string    `json:"file"`
	Function   string    `json:"function"`
	IssueType  string    `json:"issue_type"`
	Severity   string    `json:"severity"`
	LineNumber int       `json:"line_number"`
	Timestamp  string    `json:"timestamp"`
	Message    string    `json:"message"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// RiskAlert represents incoming alert from monitor
type RiskAlert struct {
	File       string `json:"file"`
	Function   string `json:"function"`
	IssueType  string `json:"issue_type"`
	Severity   string `json:"severity"`
	LineNumber int    `json:"line_number"`
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message"`
}

// MCPServer manages risk tickets in memory
type MCPServer struct {
	tickets   map[string]*RiskTicket
	mu        sync.RWMutex
	ticketSeq int
}

// MCPResponse represents the MCP protocol response
type MCPResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in MCP protocol
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPRequest represents the MCP protocol request
type MCPRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

var server *MCPServer

func init() {
	server = &MCPServer{
		tickets:   make(map[string]*RiskTicket),
		ticketSeq: 1,
	}
}

func main() {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8080"
	}

	// HTTP endpoints for receiving alerts and querying tickets
	http.HandleFunc("/api/risk-alert", handleRiskAlert)
	http.HandleFunc("/api/tickets", handleGetTickets)
	http.HandleFunc("/api/tickets/", handleGetTicket)
	
	// MCP protocol endpoint (stdio-based for IBM Bob integration)
	http.HandleFunc("/mcp", handleMCPRequest)

	log.Printf("MCP Server starting on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  POST   /api/risk-alert  - Receive risk alerts from monitor")
	log.Printf("  GET    /api/tickets     - List all risk tickets")
	log.Printf("  GET    /api/tickets/:id - Get specific ticket")
	log.Printf("  POST   /mcp             - MCP protocol endpoint for IBM Bob")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleRiskAlert receives alerts from the monitor or watsonx Orchestrate
func handleRiskAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var alert RiskAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		log.Printf("Error decoding alert: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a new ticket
	ticket := server.createTicket(alert)
	
	log.Printf("✓ New risk ticket created: %s", ticket.ID)
	log.Printf("  File: %s:%d", ticket.File, ticket.LineNumber)
	log.Printf("  Issue: %s (Severity: %s)", ticket.IssueType, ticket.Severity)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"message":   "Risk alert received and ticket created",
		"ticket_id": ticket.ID,
	})
}

// handleGetTickets returns all tickets
func handleGetTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tickets := server.getAllTickets()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"count":   len(tickets),
		"tickets": tickets,
	})
}

// handleGetTicket returns a specific ticket
func handleGetTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ticket ID from URL path
	ticketID := r.URL.Path[len("/api/tickets/"):]
	
	ticket := server.getTicket(ticketID)
	if ticket == nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"ticket": ticket,
	})
}

// handleMCPRequest handles MCP protocol requests from IBM Bob
func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendMCPError(w, nil, -32700, "Parse error")
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendMCPError(w, nil, -32700, "Parse error")
		return
	}

	// Handle different MCP methods
	switch req.Method {
	case "tools/list":
		sendMCPResponse(w, req.ID, map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "get_risk_tickets",
					"description": "Retrieve all current risk tickets from the observability system",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"severity": map[string]interface{}{
								"type":        "string",
								"description": "Filter by severity (critical, high, medium, low)",
								"enum":        []string{"critical", "high", "medium", "low"},
							},
							"status": map[string]interface{}{
								"type":        "string",
								"description": "Filter by status (open, in_progress, resolved)",
								"enum":        []string{"open", "in_progress", "resolved"},
							},
						},
					},
				},
			},
		})

	case "tools/call":
		toolName, _ := req.Params["name"].(string)
		if toolName == "get_risk_tickets" {
			tickets := server.getAllTickets()
			
			// Apply filters if provided
			if args, ok := req.Params["arguments"].(map[string]interface{}); ok {
				if severity, ok := args["severity"].(string); ok {
					tickets = filterBySeverity(tickets, severity)
				}
				if status, ok := args["status"].(string); ok {
					tickets = filterByStatus(tickets, status)
				}
			}

			sendMCPResponse(w, req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": formatTicketsForBob(tickets),
					},
				},
			})
		} else {
			sendMCPError(w, req.ID, -32601, "Method not found")
		}

	case "resources/list":
		sendMCPResponse(w, req.ID, map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"uri":         "risk://tickets",
					"name":        "Risk Tickets",
					"description": "Current risk tickets from observability monitoring",
					"mimeType":    "application/json",
				},
			},
		})

	case "resources/read":
		uri, _ := req.Params["uri"].(string)
		if uri == "risk://tickets" {
			tickets := server.getAllTickets()
			ticketsJSON, _ := json.MarshalIndent(tickets, "", "  ")
			sendMCPResponse(w, req.ID, map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      "risk://tickets",
						"mimeType": "application/json",
						"text":     string(ticketsJSON),
					},
				},
			})
		} else {
			sendMCPError(w, req.ID, -32602, "Invalid resource URI")
		}

	default:
		sendMCPError(w, req.ID, -32601, "Method not found")
	}
}

func sendMCPResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
	})
}

func sendMCPError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	})
}

func (s *MCPServer) createTicket(alert RiskAlert) *RiskTicket {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticketID := fmt.Sprintf("RISK-%d", s.ticketSeq)
	s.ticketSeq++

	ticket := &RiskTicket{
		ID:         ticketID,
		File:       alert.File,
		Function:   alert.Function,
		IssueType:  alert.IssueType,
		Severity:   alert.Severity,
		LineNumber: alert.LineNumber,
		Timestamp:  alert.Timestamp,
		Message:    alert.Message,
		Status:     "open",
		CreatedAt:  time.Now(),
	}

	s.tickets[ticketID] = ticket
	return ticket
}

func (s *MCPServer) getAllTickets() []*RiskTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tickets := make([]*RiskTicket, 0, len(s.tickets))
	for _, ticket := range s.tickets {
		tickets = append(tickets, ticket)
	}
	return tickets
}

func (s *MCPServer) getTicket(id string) *RiskTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tickets[id]
}

func filterBySeverity(tickets []*RiskTicket, severity string) []*RiskTicket {
	filtered := make([]*RiskTicket, 0)
	for _, ticket := range tickets {
		if ticket.Severity == severity {
			filtered = append(filtered, ticket)
		}
	}
	return filtered
}

func filterByStatus(tickets []*RiskTicket, status string) []*RiskTicket {
	filtered := make([]*RiskTicket, 0)
	for _, ticket := range tickets {
		if ticket.Status == status {
			filtered = append(filtered, ticket)
		}
	}
	return filtered
}

func formatTicketsForBob(tickets []*RiskTicket) string {
	if len(tickets) == 0 {
		return "No risk tickets found."
	}

	result := fmt.Sprintf("Found %d risk ticket(s):\n\n", len(tickets))
	for _, ticket := range tickets {
		result += fmt.Sprintf("**%s** [%s]\n", ticket.ID, ticket.Severity)
		result += fmt.Sprintf("- File: [`%s`](%s:%d)\n", ticket.File, ticket.File, ticket.LineNumber)
		result += fmt.Sprintf("- Function: `%s()`\n", ticket.Function)
		result += fmt.Sprintf("- Issue: %s\n", ticket.IssueType)
		result += fmt.Sprintf("- Message: %s\n", ticket.Message)
		result += fmt.Sprintf("- Status: %s\n", ticket.Status)
		result += fmt.Sprintf("- Created: %s\n\n", ticket.CreatedAt.Format(time.RFC3339))
	}
	return result
}

// Made with Bob
