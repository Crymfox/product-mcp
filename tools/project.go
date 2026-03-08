package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	product "github.com/kidkuddy/product-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterProjectTools registers tools related to the core PRODUCT.md file,
// including initialization, summaries, vision updates, and tech stack/scope management.
func RegisterProjectTools(s *server.MCPServer, projectPath string) {
	// project_init
	s.AddTool(mcp.NewTool("project_init",
		mcp.WithDescription("Initialize a new PRODUCT.md project structure"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Project description")),
		mcp.WithString("vision", mcp.Required(), mcp.Description("Initial vision statement")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := mcp.ParseString(request, "name", "")
		desc := mcp.ParseString(request, "description", "")
		vision := mcp.ParseString(request, "vision", "")

		root := projectPath
		if root == "" {
			root, _ = os.Getwd()
		}

		// Check if already exists
		if _, err := os.Stat(filepath.Join(root, "PRODUCT.md")); err == nil {
			return mcp.NewToolResultError("PRODUCT.md already exists"), nil
		}

		// Create directories
		if err := os.MkdirAll(filepath.Join(root, "domains"), 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create dirs: %v", err)), nil
		}

		// Create initial Product struct
		p := &product.Product{
			Schema:      "v1",
			Name:        name,
			Description: desc,
			Version:     "0.1.0",
			Vision:      vision,
			Goals:       []product.Goal{},
			TechStack:   []product.TechRow{},
			Scopes:      []product.Scope{},
			Domains:     []product.Domain{},
		}

		// Save PRODUCT.md (we need to temporarily create a Project wrapper or just use product.Save directly if possible)
		// product.Save is a method on *Product, taking dir.
		if err := p.Save(root); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save PRODUCT.md: %v", err)), nil
		}

		// Create empty ISSUES.md
		// We can use product.SaveIssues with empty slice
		if err := product.SaveIssues(root, []product.Issue{}); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to init ISSUES.md: %v", err)), nil
		}

		return mcp.NewToolResultText("Project initialized successfully"), nil
	})

	// project_summary
	s.AddTool(mcp.NewTool("project_summary",
		mcp.WithDescription("Get high-level project summary"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		completedGoals := 0
		for _, g := range p.Product.Goals {
			if g.Done {
				completedGoals++
			}
		}

		openIssues := len(p.OpenIssues())

		summary := fmt.Sprintf("Project: %s (v%s)\nDescription: %s\nVision: %s\n\nStats:\n- Goals: %d/%d completed\n- Open Issues: %d\n- Domains: %d\n",
			p.Product.Name, p.Product.Version, p.Product.Description, p.Product.Vision,
			completedGoals, len(p.Product.Goals), openIssues, len(p.Product.Domains))

		return mcp.NewToolResultText(summary), nil
	})

	// update_vision
	s.AddTool(mcp.NewTool("update_vision",
		mcp.WithDescription("Update the project vision"),
		mcp.WithString("vision", mcp.Required(), mcp.Description("New vision statement")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vision := mcp.ParseString(request, "vision", "")
		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		p.Product.Vision = vision
		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText("Vision updated"), nil
	})

	// manage_tech_stack
	s.AddTool(mcp.NewTool("manage_tech_stack",
		mcp.WithDescription("Add or update a technology in the tech stack"),
		mcp.WithString("layer", mcp.Required(), mcp.Description("Layer (e.g. 'Frontend', 'Database')")),
		mcp.WithString("tech", mcp.Required(), mcp.Description("Technology name")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		layer := mcp.ParseString(request, "layer", "")
		tech := mcp.ParseString(request, "tech", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		found := false
		for i := range p.Product.TechStack {
			if strings.EqualFold(p.Product.TechStack[i].Layer, layer) {
				p.Product.TechStack[i].Technology = tech
				found = true
				break
			}
		}
		if !found {
			p.Product.TechStack = append(p.Product.TechStack, product.TechRow{Layer: layer, Technology: tech})
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Tech Stack updated: %s -> %s", layer, tech)), nil
	})

	// manage_scopes
	s.AddTool(mcp.NewTool("manage_scopes",
		mcp.WithDescription("Add or update a scope in the scopes table"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Scope name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path")),
		mcp.WithString("type", mcp.Required(), mcp.Description("library, entrypoint, service, etc.")),
		mcp.WithString("state", mcp.Required(), mcp.Description("vision, building, ready, broken")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := mcp.ParseString(request, "name", "")
		path := mcp.ParseString(request, "path", "")
		typ := mcp.ParseString(request, "type", "")
		state := mcp.ParseString(request, "state", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		newScope := product.Scope{Name: name, Path: path, Type: typ, State: state}
		found := false
		for i := range p.Product.Scopes {
			if strings.EqualFold(p.Product.Scopes[i].Name, name) {
				p.Product.Scopes[i] = newScope
				found = true
				break
			}
		}
		if !found {
			p.Product.Scopes = append(p.Product.Scopes, newScope)
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Scope '%s' updated", name)), nil
	})
}
