package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	product "github.com/kidkuddy/product-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterDomainTools registers tools for managing project domains and features,
// supporting both inline definitions and external domain files.
func RegisterDomainTools(s *server.MCPServer, projectPath string) {
	// list_domains
	s.AddTool(mcp.NewTool("list_domains",
		mcp.WithDescription("List all project domains"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Domains for %s:\n\n", p.Product.Name))
		for _, d := range p.Product.Domains {
			loc := "inline"
			if path := findDomainFile(projectPath, d.Name); path != "" {
				loc = fmt.Sprintf("file: %s", filepath.Base(path))
			}
			sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", d.Name, loc, d.Summary))
			sb.WriteString(fmt.Sprintf("  Features: %d\n", len(d.Features)))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// add_domain
	s.AddTool(mcp.NewTool("add_domain",
		mcp.WithDescription("Add a new domain"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Domain name")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("One sentence summary")),
		mcp.WithBoolean("as_file", mcp.Required(), mcp.Description("If true, creates a separate domains/*.md file")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := mcp.ParseString(request, "name", "")
		summary := mcp.ParseString(request, "summary", "")
		asFile := mcp.ParseBoolean(request, "as_file", false)

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		// Check duplicate
		if p.Domain(name) != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Domain '%s' already exists", name)), nil
		}

		d := product.Domain{
			Name:    name,
			Summary: summary,
		}

		if asFile {
			slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			path := filepath.Join(projectPath, "domains", slug+".md")
			if err := product.SaveDomain(path, &d); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to create domain file: %v", err)), nil
			}
			// Add to slice so SmartSave sees it
			p.Product.Domains = append(p.Product.Domains, d)
		} else {
			p.Product.Domains = append(p.Product.Domains, d)
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Domain '%s' added", name)), nil
	})

	// add_feature
	s.AddTool(mcp.NewTool("add_feature",
		mcp.WithDescription("Add a feature to a domain"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Feature name")),
		mcp.WithString("why", mcp.Required(), mcp.Description("Why this feature is needed")),
		mcp.WithString("state", mcp.Required(), mcp.Description("vision, specced, building, ready, or broken")),
		mcp.WithArray("acceptance", mcp.Description("Initial list of acceptance criteria (strings)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		domainName := mcp.ParseString(request, "domain", "")
		featName := mcp.ParseString(request, "name", "")
		why := mcp.ParseString(request, "why", "")
		state := mcp.ParseString(request, "state", "")
		
		var acceptanceItems []product.AcceptanceItem
		if acc, ok := mcp.ParseArgument(request, "acceptance", nil).([]any); ok {
			for _, item := range acc {
				if str, ok := item.(string); ok {
					acceptanceItems = append(acceptanceItems, product.AcceptanceItem{Text: str, Done: false})
				}
			}
		}

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		d := p.Domain(domainName)
		if d == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Domain '%s' not found", domainName)), nil
		}

		// Check duplicate feature
		for _, f := range d.Features {
			if strings.EqualFold(f.Name, featName) {
				return mcp.NewToolResultError(fmt.Sprintf("Feature '%s' already exists in domain '%s'", featName, domainName)), nil
			}
		}

		newFeat := product.Feature{
			Name:       featName,
			State:      state,
			Why:        why,
			Acceptance: acceptanceItems,
		}
		
		if len(newFeat.Acceptance) == 0 {
			newFeat.Acceptance = []product.AcceptanceItem{{Text: "When X, then Y", Done: false}}
		}
		
		d.Features = append(d.Features, newFeat)

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Feature '%s' added to domain '%s'", featName, domainName)), nil
	})

	// set_feature_acceptance
	s.AddTool(mcp.NewTool("set_feature_acceptance",
		mcp.WithDescription("Overwrite all acceptance criteria for a feature"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain name")),
		mcp.WithString("feature", mcp.Required(), mcp.Description("Feature name")),
		mcp.WithArray("acceptance", mcp.Required(), mcp.Description("New list of criteria (strings)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		domainName := mcp.ParseString(request, "domain", "")
		featName := mcp.ParseString(request, "feature", "")
		
		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		f := p.Feature(domainName, featName)
		if f == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Feature '%s' in domain '%s' not found", featName, domainName)), nil
		}

		var newItems []product.AcceptanceItem
		if acc, ok := mcp.ParseArgument(request, "acceptance", nil).([]any); ok {
			for _, item := range acc {
				if str, ok := item.(string); ok {
					newItems = append(newItems, product.AcceptanceItem{Text: str, Done: false})
				}
			}
		}
		f.Acceptance = newItems

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}
		return mcp.NewToolResultText("Acceptance criteria updated"), nil
	})

	// toggle_feature_acceptance
	s.AddTool(mcp.NewTool("toggle_feature_acceptance",
		mcp.WithDescription("Mark a specific acceptance criterion as done or pending"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain name")),
		mcp.WithString("feature", mcp.Required(), mcp.Description("Feature name")),
		mcp.WithString("text", mcp.Required(), mcp.Description("The exact text of the criterion")),
		mcp.WithBoolean("done", mcp.Required(), mcp.Description("True if done, false if pending")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		domainName := mcp.ParseString(request, "domain", "")
		featName := mcp.ParseString(request, "feature", "")
		text := mcp.ParseString(request, "text", "")
		done := mcp.ParseBoolean(request, "done", false)

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		f := p.Feature(domainName, featName)
		if f == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Feature '%s' in domain '%s' not found", featName, domainName)), nil
		}

		found := false
		for i := range f.Acceptance {
			if strings.EqualFold(f.Acceptance[i].Text, text) {
				f.Acceptance[i].Done = done
				found = true
				break
			}
		}

		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("Criterion '%s' not found in feature '%s'", text, featName)), nil
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}
		return mcp.NewToolResultText("Criterion toggled"), nil
	})

	// update_feature_state
	s.AddTool(mcp.NewTool("update_feature_state",
		mcp.WithDescription("Update a feature's state"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain name")),
		mcp.WithString("feature", mcp.Required(), mcp.Description("Feature name")),
		mcp.WithString("state", mcp.Required(), mcp.Description("New state")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		domainName := mcp.ParseString(request, "domain", "")
		featName := mcp.ParseString(request, "feature", "")
		state := mcp.ParseString(request, "state", "")

		p, err := getProject(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load project: %v", err)), nil
		}

		if err := p.UpdateFeatureState(domainName, featName, state); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update feature: %v", err)), nil
		}

		if err := SmartSave(p, p.Dir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save project: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Feature '%s' state updated to %s", featName, state)), nil
	})
}
