# Quick Start Guide

This guide will help you get the Intelligent Observability & Risk Agent up and running in minutes.

## Prerequisites

- Go 1.21 or higher installed
- Terminal/PowerShell access

## Step-by-Step Setup

### 1. Start the MCP Server

Open a terminal and navigate to the project directory:

```powershell
cd D:/intelligent-observability-agent/mcp-server
go run mcp_server.go
```

You should see:
```
MCP Server starting on port 8080...
Endpoints:
  POST   /api/risk-alert  - Receive risk alerts from monitor
  GET    /api/tickets     - List all risk tickets
  GET    /api/tickets/:id - Get specific ticket
  POST   /mcp             - MCP protocol endpoint for IBM Bob
```

### 2. Test the MCP Server

Open a new terminal and test the server:

```powershell
# Test health check
curl http://localhost:8080/api/tickets
```

Expected response:
```json
{
  "status": "success",
  "count": 0,
  "tickets": []
}
```

### 3. Start the Risk Monitor

Open another terminal:

```powershell
cd D:/intelligent-observability-agent/monitor
$env:WEBHOOK_URL="http://localhost:8080/api/risk-alert"
$env:MONITOR_INTERVAL="10s"
go run monitor.go
```

You should see alerts being generated and sent:
```
Starting Risk Monitor...
Webhook URL: http://localhost:8080/api/risk-alert
Monitoring interval: 10s

=== Sending Risk Alert ===
File: src/api/handlers.go
Function: ProcessRequest
Issue: high_cpu_cycles (Severity: high)
Line: 145
Message: Detected inefficient loop consuming excessive CPU cycles
✓ Alert sent successfully (Status: 200)
```

### 4. Verify Tickets are Being Created

In another terminal, check the tickets:

```powershell
curl http://localhost:8080/api/tickets
```

You should now see risk tickets in the response.

### 5. Configure IBM Bob (Optional)

To integrate with IBM Bob in VS Code:

1. Open VS Code settings
2. Navigate to IBM Bob MCP configuration
3. Add the MCP server configuration from [`bob-mcp-config.json`](mcp-server/bob-mcp-config.json:1)

Or manually add to your Bob settings:

```json
{
  "mcpServers": {
    "risk-observability": {
      "command": "go",
      "args": ["run", "mcp_server.go"],
      "cwd": "D:/intelligent-observability-agent/mcp-server",
      "env": {
        "MCP_PORT": "8080"
      }
    }
  }
}
```

### 6. Query Risks via IBM Bob

Once configured, you can ask Bob:

- "What are the current risk tickets?"
- "Show me critical severity issues"
- "List all performance problems"

Bob will fetch the data from the MCP server and display it with clickable file links.

## Testing the Full Flow

### Manual Alert Test

Send a test alert directly to the MCP server:

```powershell
$body = @{
    file = "src/test.go"
    function = "TestFunction"
    issue_type = "high_cpu_cycles"
    severity = "critical"
    line_number = 42
    timestamp = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
    message = "Manual test alert"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/risk-alert" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
```

### Verify the Ticket

```powershell
curl http://localhost:8080/api/tickets
```

## Building for Production

Build the executables:

```powershell
# Windows
.\build.ps1

# Linux/Mac
chmod +x build.sh
./build.sh
```

Run the built binaries:

```powershell
# Start MCP Server
.\bin\mcp-server.exe

# Start Risk Monitor (in another terminal)
$env:WEBHOOK_URL="http://localhost:8080/api/risk-alert"
.\bin\risk-monitor.exe
```

## Troubleshooting

### Port Already in Use

If port 8080 is already in use, change it:

```powershell
$env:MCP_PORT="8081"
go run mcp_server.go
```

Then update the monitor's webhook URL:

```powershell
$env:WEBHOOK_URL="http://localhost:8081/api/risk-alert"
```

### Monitor Can't Connect

1. Verify MCP server is running: `curl http://localhost:8080/api/tickets`
2. Check firewall settings
3. Verify the `WEBHOOK_URL` environment variable is set correctly

### No Tickets Appearing

1. Check monitor logs for errors
2. Verify alerts are being sent (look for "✓ Alert sent successfully")
3. Check MCP server logs for incoming requests

## Next Steps

1. **Configure watsonx Orchestrate**: Import [`watsonx_tool.yaml`](orchestrate/watsonx_tool.yaml:1) to create a custom tool
2. **Customize Alerts**: Modify the mock scenarios in [`monitor.go`](monitor/monitor.go:24) to match your needs
3. **Add Persistence**: Extend the MCP server to store tickets in a database
4. **Set Up Monitoring**: Configure the monitor interval and webhook for your environment

## Support

For detailed documentation, see [`README.md`](README.md:1)

For issues or questions, please open an issue in the project repository.