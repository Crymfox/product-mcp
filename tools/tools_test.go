package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterTools(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")
	
	// Test that registration doesn't panic
	RegisterProjectTools(s, ".")
	RegisterGoalTools(s, ".")
	RegisterDomainTools(s, ".")
	RegisterIssueTools(s, ".")

	tools := s.ListTools()
	if len(tools) == 0 {
		t.Error("No tools were registered")
	}
}
