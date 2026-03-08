package tools

import (
	"context"
	"fmt"
	"strings"

	product "github.com/kidkuddy/product-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterGoalTools registers tools for managing project goals,
// including listing, adding, and toggling completion status.
func RegisterGoalTools(s *server.MCPServer, projectPath string) {
	// list_goals
	s.AddTool(mcp.NewTool("list_goals",
		mcp.WithDescription("List all project goals and their status"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Goals for %s:\n\n", p.Product.Name))
		for i, g := range p.Product.Goals {
			status := "[ ]"
			if g.Done {
				status = "[x]"
			}
			sb.WriteString(fmt.Sprintf("%d. %s **%s**: %s\n", i+1, status, g.Slug, g.Description))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// add_goal
	s.AddTool(mcp.NewTool("add_goal",
		mcp.WithDescription("Add a new goal to the project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Short identifier (e.g. 'high-performance')")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Full description of the goal")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug := mcp.ParseString(request, "slug", "")
		desc := mcp.ParseString(request, "description", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		// Check for duplicate slug
		for _, g := range p.Product.Goals {
			if strings.EqualFold(g.Slug, slug) {
				return mcp.NewToolResultError(fmt.Sprintf("Goal with slug '%s' already exists", slug)), nil
			}
		}

		p.Product.Goals = append(p.Product.Goals, product.Goal{
			Slug:        slug,
			Description: desc,
			Done:        false,
		})

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Added goal: %s", slug)), nil
	})

	// toggle_goal
	s.AddTool(mcp.NewTool("toggle_goal",
		mcp.WithDescription("Mark a goal as completed or pending"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("The slug of the goal to toggle")),
		mcp.WithBoolean("done", mcp.Required(), mcp.Description("True for completed, false for pending")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug := mcp.ParseString(request, "slug", "")
		done := mcp.ParseBoolean(request, "done", false)

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		found := false
		for i := range p.Product.Goals {
			if strings.EqualFold(p.Product.Goals[i].Slug, slug) {
				p.Product.Goals[i].Done = done
				found = true
				break
			}
		}

		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("Goal '%s' not found", slug)), nil
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		status := "pending"
		if done {
			status = "completed"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Goal '%s' marked as %s", slug, status)), nil
	})
}
