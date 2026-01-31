// Package deploy 提供 CI/CD 配置生成。
package deploy

import (
	"fmt"
	"os"
	"path/filepath"
)

// CIConfig CI/CD 配置
type CIConfig struct {
	Name           string
	Description    string
	GoVersion      string
	DockerImage    string
	DockerfilePath string
	MainPath       string
	TestTimeout    string
	EnableDocker   bool
	EnableK8s      bool
}

// DefaultCIConfig 默认 CI 配置
func DefaultCIConfig(name string) *CIConfig {
	return &CIConfig{
		Name:           name,
		Description:    "Shode application CI/CD pipeline",
		GoVersion:      "1.21",
		DockerImage:    fmt.Sprintf("%s:latest", name),
		DockerfilePath: "Dockerfile",
		MainPath:       "cmd/main.go",
		TestTimeout:    "10m",
		EnableDocker:   true,
		EnableK8s:      false,
	}
}

// CIGenerator CI 配置生成器
type CIGenerator struct {
	config *CIConfig
}

// NewCIGenerator 创建 CI 生成器
func NewCIGenerator(config *CIConfig) *CIGenerator {
	return &CIGenerator{config: config}
}

// GenerateGitHubActions 生成 GitHub Actions 工作流
func (g *CIGenerator) GenerateGitHubActions() string {
	var dockerBuild string
	var dockerPush string
	var k8sDeploy string

	if g.config.EnableDocker {
		dockerBuild = fmt.Sprintf(`      -
        name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      -
        name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      -
        name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: %s
          push: true
          tags: ${{ secrets.DOCKER_USERNAME }}/%s:${{ github.ref_name }}
          tags: ${{ secrets.DOCKER_USERNAME }}/%s:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
`,
			g.config.DockerfilePath,
			g.config.Name,
			g.config.Name,
		)
	}

	if g.config.EnableK8s {
		k8sDeploy = fmt.Sprintf(`  deploy:
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/main'
    steps:
      -
        name: Checkout code
        uses: actions/checkout@v4

      -
        name: Set up kubectl
        uses: azure/setup-kubectl@v3

      -
        name: Deploy to Kubernetes
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig
          kubectl apply -f k8s/
`)
	}

	return fmt.Sprintf(`name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: "%s"
  IMAGE_NAME: %s

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      -
        name: Checkout code
        uses: actions/checkout@v4

      -
        name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      -
        name: Download dependencies
        run: go mod download

      -
        name: Verify dependencies
        run: go mod verify

      -
        name: Run go vet
        run: go vet ./...

      -
        name: Run tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      -
        name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  lint:
    runs-on: ubuntu-latest
    steps:
      -
        name: Checkout code
        uses: actions/checkout@v4

      -
        name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      -
        name: Checkout code
        uses: actions/checkout@v4

      -
        name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      -
        name: Download dependencies
        run: go mod download

      -
        name: Build
        run: |
          CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app %s

      -
        name: Upload build artifact
        uses: actions/upload-artifact@v3
        with:
          name: app
          path: app
%s
%s
`,
		g.config.GoVersion,
		g.config.DockerImage,
		g.config.MainPath,
		dockerBuild,
		k8sDeploy,
	)
}

// GenerateGitLabCI 生成 GitLab CI 配置
func (g *CIGenerator) GenerateGitLabCI() string {
	var dockerBuild string
	if g.config.EnableDocker {
		dockerBuild = fmt.Sprintf(`
docker:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - docker login -u "$CI_REGISTRY_USER" -p "$CI_REGISTRY_PASSWORD" $CI_REGISTRY
  script:
    - docker build -t %s -f %s .
    - docker push %s
  only:
    - main
`,
			g.config.DockerImage,
			g.config.DockerfilePath,
			g.config.DockerImage,
		)
	}

	return fmt.Sprintf(`image: golang:%s

stages:
  - test
  - build
  - deploy

variables:
  GO_TESTS: "./..."
  GO_TEST_TIMEOUT: "%s"

before_script:
  - mkdir -p .go
  - go mod download

test:
  stage: test
  script:
    - go vet ./...
    - go test -v -race -coverprofile=coverage.out $GO_TESTS
  coverage: '/coverage: \d+.\d+\% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out

lint:
  stage: test
  image: golangci/golangci-lint:latest
  script:
    - golangci-lint run ./...

build:
  stage: build
  script:
    - go build -o app %s
  artifacts:
    paths:
      - app
%s
`,
		g.config.GoVersion,
		g.config.TestTimeout,
		g.config.MainPath,
		dockerBuild,
	)
}

// GenerateJenkinsfile 生成 Jenkins 配置
func (g *CIGenerator) GenerateJenkinsfile() string {
	return fmt.Sprintf(`pipeline {
    agent any

    tools {
        go '%s'
    }

    environment {
        APP_NAME = '%s'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Install Dependencies') {
            steps {
                sh 'go mod download'
                sh 'go mod verify'
            }
        }

        stage('Vet') {
            steps {
                sh 'go vet ./...'
            }
        }

        stage('Test') {
            steps {
                sh 'go test -v -race -coverprofile=coverage.out -covermode=atomic ./...'
            }
            post {
                always {
                    recordCoverage(
                        tools: [[parser: 'GoCoverage']]
                    )
                }
            }
        }

        stage('Build') {
            steps {
                sh 'go build -o ${APP_NAME} %s'
            }
        }

%s        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                echo 'Deploying to production...'
                // Add your deployment commands here
            }
        }
    }

    post {
        always {
            cleanWs()
        }
        success {
            echo 'Pipeline succeeded!'
        }
        failure {
            echo 'Pipeline failed!'
        }
    }
}
`,
		g.config.GoVersion,
		g.config.Name,
		g.config.MainPath,
		g.getJenkinsDockerStage(),
	)
}

// getJenkinsDockerStage 获取 Jenkins Docker 阶段
func (g *CIGenerator) getJenkinsDockerStage() string {
	if !g.config.EnableDocker {
		return ""
	}

	return fmt.Sprintf(`        stage('Docker Build') {
            when {
                branch 'main'
            }
            steps {
                script {
                    docker.build("%s", "-f %s .")
                }
            }
        }

`)
}

// GenerateMakefile 生成 Makefile
func (g *CIGenerator) GenerateMakefile() string {
	return fmt.Sprintf(`.PHONY: all build test lint clean docker-build docker-push run

APP_NAME=%s
MAIN_PATH=%s
DOCKER_IMAGE=%s
DOCKERFILE=%s
GO_VERSION=%s

all: build

build:
	go build -o $(APP_NAME) $(MAIN_PATH)

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...
	golangci-lint run ./...

clean:
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html

docker-build:
	docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .

docker-push:
	docker push $(DOCKER_IMAGE)

run: build
	./$(APP_NAME)

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

deps-update:
	go get -u ./...
	go mod tidy
`,
		g.config.Name,
		g.config.MainPath,
		g.config.DockerImage,
		g.config.DockerfilePath,
		g.config.GoVersion,
	)
}

// GenerateAll 生成所有 CI/CD 配置
func (g *CIGenerator) GenerateAll() map[string]string {
	files := make(map[string]string)

	// GitHub Actions
	files[fmt.Sprintf(".github/workflows/ci.yml")] = g.GenerateGitHubActions()

	// GitLab CI
	files[".gitlab-ci.yml"] = g.GenerateGitLabCI()

	// Jenkins
	files["Jenkinsfile"] = g.GenerateJenkinsfile()

	// Makefile
	files["Makefile"] = g.GenerateMakefile()

	return files
}

// WriteCIConfigs 写入 CI/CD 配置文件
func WriteCIConfigs(config *CIConfig, outputDir string) error {
	generator := NewCIGenerator(config)
	files := generator.GenerateAll()

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 写入文件
	for filepath, content := range files {
		fullPath := filepath.Join(outputDir, filepath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filepath, err)
		}
	}

	return nil
}
