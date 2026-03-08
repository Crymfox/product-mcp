package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	product "github.com/kidkuddy/product-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterIssueTools registers tools for managing the project backlog,
// including creating, updating, and filtering issues in ISSUES.md.
func RegisterIssueTools(s *server.MCPServer, projectPath string) {
	// list_issues
	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List issues with optional status filter"),
		mcp.WithString("status", mcp.Description("Filter by status (e.g. 'open', 'closed')")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status := mcp.ParseString(request, "status", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		var issues []product.Issue
		if status != "" {
			issues = p.IssuesByStatus(status)
		} else {
			issues = p.Issues
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(issues)))
		for _, iss := range issues {
			sb.WriteString(fmt.Sprintf("- **[%s]** %s (Status: %s, Severity: %s)\n", iss.ID, iss.Title, iss.Status, iss.Severity))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// create_issue
	s.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new issue"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Issue title")),
		mcp.WithString("type", mcp.Required(), mcp.Description("bug, task, or improvement")),
		mcp.WithString("severity", mcp.Required(), mcp.Description("low, medium, high, or critical")),
		mcp.WithString("effort", mcp.Required(), mcp.Description("Estimated effort (e.g. '2h', '3d')")),
		mcp.WithString("body", mcp.Description("Detailed description")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := mcp.ParseString(request, "title", "")
		typ := mcp.ParseString(request, "type", "")
		sev := mcp.ParseString(request, "severity", "")
		effort := mcp.ParseString(request, "effort", "")
		body := mcp.ParseString(request, "body", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		// Calculate next ID
		maxID := 0
		for _, iss := range p.Issues {
			// ID is "ISSUE-NNN"
			if strings.HasPrefix(iss.ID, "ISSUE-") {
				numStr := strings.TrimPrefix(iss.ID, "ISSUE-")
				if num, err := strconv.Atoi(numStr); err == nil {
					if num > maxID {
						maxID = num
					}
				}
			}
		}
		newID := fmt.Sprintf("ISSUE-%03d", maxID+1)

		issue := product.Issue{
			ID:       newID,
			Title:    title,
			Type:     typ,
			Severity: sev,
			Status:   "open",
			Effort:   effort,
			Body:     body,
		}

		if err := p.AddIssue(issue); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add issue: %v", err)), nil
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Created issue %s", newID)), nil
	})

	// update_issue
	s.AddTool(mcp.NewTool("update_issue",
		mcp.WithDescription("Update an issue"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Issue ID (e.g. ISSUE-001)")),
		mcp.WithString("status", mcp.Description("New status")),
		mcp.WithString("fix", mcp.Description("Description of the fix (if closing)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(request, "id", "")
		status := mcp.ParseString(request, "status", "")
		fix := mcp.ParseString(request, "fix", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		err = p.UpdateIssue(id, func(iss *product.Issue) {
			if status != "" {
				iss.Status = status
			}
			if fix != "" {
				iss.Fix = fix
			}
		})

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update issue: %v", err)), nil
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Updated issue %s", id)), nil
	})
}
