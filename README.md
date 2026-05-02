# Intelligent Observability & Risk Agent

A comprehensive observability and risk management system that integrates Go-based monitoring, watsonx Orchestrate, and IBM Bob via Model Context Protocol (MCP).

## Architecture Overview

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────┐
│  Risk Monitor   │────────▶│ watsonx          │────────▶│ MCP Server  │
│  (Go eBPF Mock) │  HTTP   │ Orchestrate      │  HTTP   │  (Go)       │
└─────────────────┘         └──────────────────┘         └─────────────┘
                                                                  │
                                                                  │ MCP
                                                                  ▼
                                                           ┌─────────────┐
                                                           │  IBM Bob    │
                                                           │  (VS Code)  │
                                                           └─────────────┘
```

## Components

### 1. Risk Monitor (`monitor/`)
A lightweight Go application that simulates an eBPF performance and carbon monitoring probe.

**Features:**
- Generates mock risk alerts for various issues (CPU cycles, memory leaks, inefficient algorithms, carbon footprint)
- Sends JSON payloads via HTTP POST to configurable webhook
- Configurable monitoring intervals
- Multiple severity levels (critical, high, medium, low)

**Configuration:**
- `WEBHOOK_URL`: Target endpoint (default: `http://localhost:8080/api/risk-alert`)
- `MONITOR_INTERVAL`: Monitoring interval (default: `30s`)

**Run:**
```bash
cd monitor
go run monitor.go
```

**Build:**
```bash
cd monitor
go build -o risk-monitor monitor.go
```

### 2. watsonx Orchestrate Tool (`orchestrate/`)
OpenAPI 3.0 specification for creating a custom tool in watsonx Orchestrate.

**Features:**
- Receives risk alerts from the monitor
- Validates payload structure
- Forwards alerts to MCP server
- Supports multiple issue types and severity levels

**Import to watsonx Orchestrate:**
1. Log in to your watsonx Orchestrate instance
2. Navigate to Tools → Import Tool
3. Upload `watsonx_tool.yaml`
4. Configure authentication (Bearer token)
5. Update the server URL to your actual endpoint

### 3. MCP Server (`mcp-server/`)
Model Context Protocol server that stores risk tickets and exposes them to IBM Bob.

**Features:**
- Receives risk alerts via HTTP POST
- Stores tickets in memory with unique IDs
- Exposes MCP protocol endpoints for IBM Bob integration
- Provides REST API for ticket management
- Supports filtering by severity and status

**Configuration:**
- `MCP_PORT`: Server port (default: `8080`)

**Endpoints:**
- `POST /api/risk-alert` - Receive risk alerts
- `GET /api/tickets` - List all tickets
- `GET /api/tickets/:id` - Get specific ticket
- `POST /mcp` - MCP protocol endpoint for IBM Bob

**Run:**
```bash
cd mcp-server
go run mcp_server.go
```

**Build:**
```bash
cd mcp-server
go build -o mcp-server mcp_server.go
```

## Setup Instructions

### Prerequisites
- Go 1.21 or higher
- watsonx Orchestrate account
- IBM Bob (VS Code extension)

### Step 1: Start the MCP Server
```bash
cd mcp-server
go run mcp_server.go
```

The server will start on port 8080 by default.

### Step 2: Configure watsonx Orchestrate
1. Import `orchestrate/watsonx_tool.yaml` into watsonx Orchestrate
2. Update the server URL in the OpenAPI spec to point to your MCP server
3. Configure authentication credentials
4. Test the tool to ensure it can receive alerts

### Step 3: Configure IBM Bob MCP Connection
Add the following to your IBM Bob MCP configuration:

```json
{
  "mcpServers": {
    "risk-observability": {
      "url": "http://localhost:8080/mcp",
      "transport": "http"
    }
  }
}
```

### Step 4: Start the Risk Monitor
```bash
cd monitor
export WEBHOOK_URL="http://localhost:8080/api/risk-alert"
export MONITOR_INTERVAL="30s"
go run monitor.go
```

For production, update `WEBHOOK_URL` to point to your watsonx Orchestrate endpoint.

## Usage

### Querying Risk Tickets via IBM Bob

Once the system is running, you can ask IBM Bob about current risks:

**Example queries:**
- "What are the current risk tickets?"
- "Show me all critical severity issues"
- "List open risk tickets"
- "What performance issues have been detected?"

IBM Bob will use the MCP server to fetch and display the risk tickets with clickable file links.

### Direct API Access

You can also query the MCP server directly:

```bash
# List all tickets
curl http://localhost:8080/api/tickets

# Get specific ticket
curl http://localhost:8080/api/tickets/RISK-1

# Manually send a risk alert
curl -X POST http://localhost:8080/api/risk-alert \
  -H "Content-Type: application/json" \
  -d '{
    "file": "src/test.go",
    "function": "TestFunction",
    "issue_type": "high_cpu_cycles",
    "severity": "high",
    "line_number": 42,
    "timestamp": "2026-05-02T03:00:00Z",
    "message": "Test alert"
  }'
```

## Risk Alert Schema

```json
{
  "file": "src/api/handlers.go",
  "function": "ProcessRequest",
  "issue_type": "high_cpu_cycles",
  "severity": "high",
  "line_number": 145,
  "timestamp": "2026-05-02T03:00:00Z",
  "message": "Detected inefficient loop consuming excessive CPU cycles"
}
```

**Issue Types:**
- `high_cpu_cycles` - Excessive CPU consumption
- `memory_leak` - Potential memory leak
- `inefficient_algorithm` - Suboptimal algorithm complexity
- `high_carbon_footprint` - High carbon emissions from compute
- `blocking_io` - Blocking I/O operations

**Severity Levels:**
- `critical` - Immediate attention required
- `high` - High priority
- `medium` - Medium priority
- `low` - Low priority

## Development

### Running Tests
```bash
# Test monitor
cd monitor
go test ./...

# Test MCP server
cd mcp-server
go test ./...
```

### Building for Production
```bash
# Build all components
./build.sh

# Or build individually
cd monitor && go build -o ../bin/risk-monitor monitor.go
cd mcp-server && go build -o ../bin/mcp-server mcp_server.go
```

## Troubleshooting

### Monitor not sending alerts
- Check `WEBHOOK_URL` is correctly set
- Verify MCP server is running and accessible
- Check monitor logs for connection errors

### IBM Bob not showing tickets
- Verify MCP server is running on the configured port
- Check IBM Bob MCP configuration
- Ensure the MCP endpoint is accessible from VS Code

### watsonx Orchestrate integration issues
- Verify OpenAPI spec is correctly imported
- Check authentication credentials
- Ensure server URL is accessible from watsonx Orchestrate

## License

MIT License

## Contributing

Contributions are welcome! Please submit pull requests or open issues for bugs and feature requests.