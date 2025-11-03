#!/bin/bash
# Integration test script for openplantbook-mcp
# Tests basic functionality without full MCP protocol testing

set -e

echo "=== OpenPlantbook MCP Server Integration Tests ==="
echo

# Load environment variables
if [ -f .env ]; then
    source .env
    echo "✓ Loaded API credentials from .env"
else
    echo "✗ No .env file found"
    exit 1
fi

# Check API key is set
if [ -z "$OPENPLANTBOOK_API_KEY" ]; then
    echo "✗ OPENPLANTBOOK_API_KEY not set"
    exit 1
fi
echo "✓ API credentials configured"
echo

# Build the binary
echo "Building openplantbook-mcp..."
make build > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Build successful"
else
    echo "✗ Build failed"
    exit 1
fi
echo

# Test version command
echo "Testing version command..."
VERSION_OUTPUT=$(./bin/openplantbook-mcp --version 2>&1)
if echo "$VERSION_OUTPUT" | grep -q "openplantbook-mcp"; then
    echo "✓ Version command works"
    echo "  $VERSION_OUTPUT" | head -1
else
    echo "✗ Version command failed"
    exit 1
fi
echo

# Test that server can start (just check it doesn't error on config)
echo "Testing server configuration..."
# Create a test that the server can initialize
# We'll use timeout to kill it after 1 second since it will wait for stdin
# Export the variable so subprocess can see it
export OPENPLANTBOOK_API_KEY
timeout 1 ./bin/openplantbook-mcp 2>&1 | head -10 > /tmp/mcp_test.log || true

if grep -q "starting openplantbook-mcp server" /tmp/mcp_test.log; then
    echo "✓ Server starts successfully and loads configuration"
elif grep -q "authentication required" /tmp/mcp_test.log; then
    echo "✗ Server configuration error (API key not loaded):"
    cat /tmp/mcp_test.log
    exit 1
elif grep -q "error" /tmp/mcp_test.log; then
    echo "✗ Server error:"
    cat /tmp/mcp_test.log
    exit 1
else
    # Timeout is expected - server is waiting for MCP input
    echo "✓ Server initializes (waiting for MCP input as expected)"
fi
echo

# Test API connectivity using SDK directly
echo "Testing API connectivity..."
TESTFILE="/tmp/apicheck.go"
cat > "$TESTFILE" << 'GOTEST'
package main

import (
    "context"
    "fmt"
    "os"
    openplantbook "github.com/rmrfslashbin/openplantbook-go"
)

func main() {
    apiKey := os.Getenv("OPENPLANTBOOK_API_KEY")
    client, err := openplantbook.New(openplantbook.WithAPIKey(apiKey))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
        os.Exit(1)
    }

    // Test search
    results, err := client.SearchPlants(context.Background(), "monstera", &openplantbook.SearchOptions{Limit: 3})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("✓ Search: Found %d plants\n", len(results))

    if len(results) == 0 {
        fmt.Fprintf(os.Stderr, "No results found\n")
        os.Exit(1)
    }

    // Test details
    pid := results[0].PID
    details, err := client.GetPlantDetails(context.Background(), pid, nil)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Get details failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("✓ Details: Retrieved %s\n", details.PID)
    fmt.Printf("  Light: %d-%d lux\n", details.MinLightLux, details.MaxLightLux)
    fmt.Printf("  Temp: %.1f-%.1f°C\n", details.MinTemp, details.MaxTemp)
}
GOTEST

cd /tmp && go mod init test 2>/dev/null || true
go get github.com/rmrfslashbin/openplantbook-go@v1.0.0 2>/dev/null
if go run apicheck.go 2>&1; then
    echo "✓ API integration successful"
else
    echo "✗ API test failed"
    exit 1
fi
echo

echo "=== All Integration Tests Passed ==="
echo
echo "The MCP server is ready to use. To test with Claude Desktop:"
echo "  1. Add configuration to Claude Desktop config file"
echo "  2. Restart Claude Desktop"
echo "  3. Ask Claude about plant care"
echo
echo "See README.md for full setup instructions."
