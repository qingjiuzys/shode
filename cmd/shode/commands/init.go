package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Project templates
const (
	templateBasic = `#!/usr/bin/env shode
# {{.ProjectName}} - Basic Shode Project

# Start HTTP server
StartHTTPServer "3000"

# Your code here

Println "Server running at http://localhost:3000"

# Keep server running
for i in $(seq 1 100000); do sleep 1; done
`

	templateStaticServer = `#!/usr/bin/env shode
# {{.ProjectName}} - Static File Server

StartHTTPServer "3000"

# Serve static files
RegisterStaticRoute "/" "./public"

Println "Static file server running at http://localhost:3000"

for i in $(seq 1 100000); do sleep 1; done
`

	templateAPI = `#!/usr/bin/env shode
# {{.ProjectName}} - RESTful API Server

StartHTTPServer "3000"

# API: Health check
function healthCheck() {
    SetHTTPResponse 200 '{"status":"healthy"}'
}
RegisterHTTPRoute "GET" "/api/health" "function" "healthCheck"

# API: Get items
function getItems() {
    SetHTTPResponse 200 '{"items":[]}'
}
RegisterHTTPRoute "GET" "/api/items" "function" "getItems"

# API: Create item
function createItem() {
    SetHTTPResponse 201 '{"id":1,"created":true}'
}
RegisterHTTPRoute "POST" "/api/items" "function" "createItem"

Println "API server running at http://localhost:3000"

for i in $(seq 1 100000); do sleep 1; done
`

	readmeContent = `# {{.ProjectName}}

A Shode project created with Shode v0.5.0.

## Getting Started

Run the project:

    shode run {{.ProjectName}}.sh

## Project Structure

{{.ProjectName}}/
  - {{.ProjectName}}.sh    # Main entry point
  - public/                # Static files (if applicable)
  - README.md             # This file

## Documentation

- Shode User Guide: https://docs.shode.818cloud.com/
- Static File Server Guide: https://docs.shode.818cloud.com/static-file-server.html
- API Reference: https://docs.shode.818cloud.com/api.html

## License

MIT
`
)

// NewInitCommand creates the 'init' command
func NewInitCommand() *cobra.Command {
	var projectType string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Create a new Shode project",
		Long: `Init creates a new Shode project with the specified name and template.

Supported project types:
  - basic:     Basic empty project
  - static:    Static file server
  - api:       RESTful API server

Examples:
  shode init myproject
  shode init myproject --type=static
  shode init website --type=api`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]

			// Validate project type
			if projectType == "" {
				projectType = "basic"
			}

			// Create project directory
			projectDir := projectName
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
				os.Exit(1)
			}

			// Generate main script
			scriptPath := filepath.Join(projectDir, projectName+".sh")
			var scriptContent string

			switch projectType {
			case "basic":
				scriptContent = templateBasic
			case "static":
				scriptContent = templateStaticServer
				// Create public directory
				publicDir := filepath.Join(projectDir, "public")
				if err := os.MkdirAll(publicDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating public directory: %v\n", err)
					os.Exit(1)
				}
				// Create index.html
				indexHTML := `<!DOCTYPE html>
<html>
<head>
    <title>` + projectName + `</title>
</head>
<body>
    <h1>Welcome to ` + projectName + `</h1>
    <p>Your Shode project is ready!</p>
</body>
</html>`
				if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(indexHTML), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating index.html: %v\n", err)
					os.Exit(1)
				}
			case "api":
				scriptContent = templateAPI
			default:
				fmt.Fprintf(os.Stderr, "Error: Unknown project type '%s'\n", projectType)
				fmt.Fprintf(os.Stderr, "Valid types: basic, static, api\n")
				os.Exit(1)
			}

			// Replace template variables
			scriptContent = replaceTemplate(scriptContent, projectName)

			// Write main script
			if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating script: %v\n", err)
				os.Exit(1)
			}

			// Generate README
			readmePath := filepath.Join(projectDir, "README.md")
			readmeData := replaceTemplate(readmeContent, projectName)
			if err := os.WriteFile(readmePath, []byte(readmeData), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating README: %v\n", err)
				os.Exit(1)
			}

			// Success message
			fmt.Printf("✓ Created project: %s\n", projectName)
			fmt.Printf("✓ Type: %s\n", projectType)
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  cd %s\n", projectName)
			fmt.Printf("  shode run %s.sh\n", projectName)
			fmt.Printf("\nEnjoy building with Shode!\n")
		},
	}

	cmd.Flags().StringVarP(&projectType, "type", "t", "basic", "Project type (basic, static, api)")

	return cmd
}

// replaceTemplate replaces template variables with actual values
func replaceTemplate(content, projectName string) string {
	// Simple template replacement
	result := content
	// In a real implementation, use text/template package
	result = replaceAll(result, "{{.ProjectName}}", projectName)
	return result
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	result := ""
	index := 0
	for {
		i := indexOf(s[index:], old)
		if i == -1 {
			result += s[index:]
			break
		}
		result += s[index:index+i] + new
		index += i + len(old)
	}
	return result
}

// indexOf finds the index of substr in s, starting from index 0
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
