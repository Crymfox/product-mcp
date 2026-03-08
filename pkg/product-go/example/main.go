// Command example demonstrates the product-go library by loading the taskr
// testdata project, querying it, mutating it, and saving a copy.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	product "github.com/kidkuddy/product-go"
)

func main() {
	// 1. Open the project from the testdata directory.
	proj, err := product.Open("testdata")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// 2. Print basic project info.
	fmt.Printf("Project:  %s\n", proj.Product.Name)
	fmt.Printf("Version:  %s\n", proj.Product.Version)
	fmt.Printf("Vision:   %s\n\n", proj.Product.Vision)

	// 3. List all open issues.
	open := proj.OpenIssues()
	fmt.Printf("Open issues (%d):\n", len(open))
	for _, iss := range open {
		fmt.Printf("  [%s] %s  (severity: %s)\n", iss.ID, iss.Title, iss.Severity)
	}
	fmt.Println()

	// 4. Mutate: mark the add-task feature as ready.
	if err := proj.UpdateFeatureState("cli", "add-task", "ready"); err != nil {
		fmt.Fprintf(os.Stderr, "UpdateFeatureState: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Updated add-task state -> ready")

	// 5. Mutate: close ISSUE-004.
	if err := proj.CloseIssue("ISSUE-004"); err != nil {
		fmt.Fprintf(os.Stderr, "CloseIssue: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Closed ISSUE-004")

	// 6. Save to a temp copy so testdata is not modified.
	tmpDir := filepath.Join(os.TempDir(), "taskr-example-output")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "MkdirAll: %v\n", err)
		os.Exit(1)
	}

	// Save product + issues to temp dir.
	if err := proj.Product.Save(tmpDir); err != nil {
		fmt.Fprintf(os.Stderr, "Product.Save: %v\n", err)
		os.Exit(1)
	}
	if err := product.SaveIssues(tmpDir, proj.Issues); err != nil {
		fmt.Fprintf(os.Stderr, "SaveIssues: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved to %s\n", tmpDir)

	// 7. Done.
	fmt.Println("Done.")
}
