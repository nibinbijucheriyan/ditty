# Build script for Intelligent Observability Agent (PowerShell)

Write-Host "Building Intelligent Observability Agent..." -ForegroundColor Green

# Create bin directory
New-Item -ItemType Directory -Path "bin" -Force | Out-Null

# Build Risk Monitor
Write-Host "Building Risk Monitor..." -ForegroundColor Cyan
Set-Location monitor
go build -o ../bin/risk-monitor.exe monitor.go
Set-Location ..

# Build MCP Server
Write-Host "Building MCP Server..." -ForegroundColor Cyan
Set-Location mcp-server
go build -o ../bin/mcp-server.exe mcp_server.go
Set-Location ..

Write-Host "`nBuild complete! Binaries are in the bin/ directory:" -ForegroundColor Green
Get-ChildItem bin/ | Format-Table Name, Length, LastWriteTime

Write-Host "`nTo run the components:" -ForegroundColor Yellow
Write-Host "  .\bin\mcp-server.exe"
Write-Host "  .\bin\risk-monitor.exe"

# Made with Bob
