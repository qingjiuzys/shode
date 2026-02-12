// Package scaffold 提供项目脚手架功能。
package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectTemplate 项目模板
type ProjectTemplate struct {
	Name        string
	Description string
	ModuleName   string
	Author      string
	License     string
	GoVersion   string
}

// NewProjectTemplate 创建项目模板
func NewProjectTemplate(name string) *ProjectTemplate {
	return &ProjectTemplate{
		Name:        name,
		Description: "A Shode web application",
		ModuleName:   strings.ToLower(strings.ReplaceAll(name, " ", "")),
		Author:      getGitUser(),
		License:     "MIT",
		GoVersion:   "1.21",
	}
}

// CreateProject 创建新项目
func (pt *ProjectTemplate) CreateProject(targetDir string) error {
	// 创建项目目录结构
	dirs := []string{
		"cmd",
		"pkg",
		"web",
		"config",
		"scripts",
		"tests",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 生成 go.mod
	if err := pt.generateGoMod(targetDir); err != nil {
		return err
	}

	// 生成 main.go
	if err := pt.generateMain(targetDir); err != nil {
		return err
	}

	// 生成 README.md
	if err := pt.generateReadme(targetDir); err != nil {
		return err
	}

	// 生成 .gitignore
	if err := pt.generateGitignore(targetDir); err != nil {
		return err
	}

	// 生成 Dockerfile
	if err := pt.generateDockerfile(targetDir); err != nil {
		return err
	}

	// 生成 docker-compose.yml
	if err := pt.generateDockerCompose(targetDir); err != nil {
		return err
	}

	fmt.Printf("✓ Project '%s' created in %s\n", pt.Name, targetDir)
	fmt.Printf("✓ Run 'cd %s && go mod tidy' to initialize dependencies\n", targetDir)

	return nil
}

func (pt *ProjectTemplate) generateGoMod(dir string) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	gitee.com/com_818cloud/shode v0.10.0
)
`, pt.ModuleName)

	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
}

func (pt *ProjectTemplate) generateMain(dir string) error {
	content := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"net/http"

	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func main() {
	// TODO: Initialize your application

	// Start HTTP server on port 8080
	stdLib := stdlib.New()

	// Register a simple route
	stdLib.RegisterRouteWithResponse("/", "Hello from %s!")

	// Start HTTP server
	stdLib.StartHTTPServer("8080")

	log.Println("Server started on http://localhost:8080")
	select {}
}
`, pt.Name)

	return os.WriteFile(filepath.Join(dir, "cmd", "main.go"), []byte(content), 0644)
}

func (pt *ProjectTemplate) generateReadme(dir string) error {
	content := fmt.Sprintf(`# %s

%s

## Getting Started

`+"```bash"+`# Install dependencies
go mod tidy

# Run the application
go run cmd/main.go
`+"```"+`

## Project Structure

`+"```"+`%s/
├── cmd/           # Command-line applications
├── pkg/           # Package code
├── web/           # Web assets
├── config/        # Configuration files
├── scripts/       # Build and deployment scripts
├── tests/         # Test files
└── docs/          # Documentation
`+"```"+`

## Configuration

Configuration files are stored in the `+"`config/`"+` directory.

## Contributing

1. Fork the repository
2. Create your feature branch (`+"`git checkout -b feature/amazing-feature`"+`)
3. Commit your changes (`+"`git commit -m 'Add some amazing feature'`"+`)
4. Push to the branch (`+"`git push origin feature/amazing-feature`"+`)
5. Open a Pull Request

## License

%s - See LICENSE file for details.
`, pt.Name, pt.Description, pt.ModuleName, pt.License)

	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0644)
}

func (pt *ProjectTemplate) generateGitignore(dir string) error {
	content := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with ` + "`go test -c`" + `
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Config
config/*.local
.env
*.local

# Build
build/
dist/
`

	return os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644)
}

func (pt *ProjectTemplate) generateDockerfile(dir string) error {
	content := fmt.Sprintf(`# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /app/main cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main /app/main
EXPOSE 8080
CMD ["/app/main"]
`)

	return os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(content), 0644)
}

func (pt *ProjectTemplate) generateDockerCompose(dir string) error {
	content := fmt.Sprintf(`version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - CONFIG_PATH=/config/config.yaml
    volumes:
      - ./config:/config:ro
    depends_on:
      - db

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=%s_db
      - POSTGRES_USER=%s_user
      - POSTGRES_PASSWORD=%s_pass
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:
`, pt.ModuleName, pt.ModuleName, pt.ModuleName)

	return os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0644)
}

func getGitUser() string {
	// 尝试从 git 获取用户名
	if user := os.Getenv("GIT_AUTHOR_NAME"); user != "" {
		return user
	}
	return "Your Name"
}

// AskUser 询问用户
func AskUser(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultValue)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		input = defaultValue
	}

	return input
}

// InteractiveCreate 交互式创建项目
func InteractiveCreate() error {
	fmt.Println("Shode Project Scaffold")
	fmt.Println("==================")
	fmt.Println()

	name := AskUser("Project name", "myapp")
	description := AskUser("Description", "My awesome Shode application")

	targetDir := AskUser("Target directory", "./"+name)

	pt := NewProjectTemplate(name)
	pt.Description = description

	return pt.CreateProject(targetDir)
}
