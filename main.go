package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/CrymfoxLabs/product-mcp/tools"
)

func main() {
	// Default to empty for discovery mode
	projectPath := ""
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	}

	// Create MCP server
	s := server.NewMCPServer(
		"product-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	// Register tools
	tools.RegisterProjectTools(s, projectPath)
	tools.RegisterGoalTools(s, projectPath)
	tools.RegisterDomainTools(s, projectPath)
	tools.RegisterIssueTools(s, projectPath)

	// Serve
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
