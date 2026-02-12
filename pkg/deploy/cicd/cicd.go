// Package cicd CI/CD 配置生成器
package cicd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CICDGenerator CI/CD 配置生成器
type CICDGenerator struct {
	platform    string // "github", "gitlab", "jenkins"
	projectName string
	language    string
}

// NewCICDGenerator 创建 CI/CD 生成器
func NewCICDGenerator(platform, projectName string) *CICDGenerator {
	return &CICDGenerator{
		platform:    platform,
		projectName: projectName,
		language:    "go",
	}
}

// Generate 生成 CI/CD 配置
func (cg *CICDGenerator) Generate() error {
	fmt.Printf("🔧 Generating %s CI/CD configuration...\n", strings.Title(cg.platform))

	switch cg.platform {
	case "github":
		return cg.generateGitHubActions()
	case "gitlab":
		return cg.generateGitLabCI()
	case "jenkins":
		return cg.generateJenkins()
	default:
		return fmt.Errorf("unsupported platform: %s", cg.platform)
	}
}

// generateGitHubActions 生成 GitHub Actions 配置
func (cg *CICDGenerator) generateGitHubActions() error {
	workflow := `name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: $\{{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: $\{{ hashFiles('**/go.sum') }}
          restore-keys: |
            $\{{ hashFiles('**/go.sum') }}

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: test

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: $\{{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: $\{{ hashFiles('**/go.sum') }}

      - name: Build binary
        run: |
          CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/shode

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: app
          path: app

  docker:
    name: Build and Push Docker Image
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Container Registry
        uses: docker/login-action@v2
        with:
          registry: $\{{ env.REGISTRY }}
          username: $\{{ github.actor }}
          password: $\{{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: |
            $\{{ env.REGISTRY }}/$\{{ github.repository }}:latest
            $\{{ env.REGISTRY }}/$\{{ github.repository }}:$\{{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    name: Deploy to Kubernetes
    runs-on: ubuntu-latest
    needs: docker
    if: github.ref == 'refs/heads/main'

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: $\{{ secrets.KUBE_CONFIG }}

      - name: Deploy to Kubernetes
        run: |
          kubectl apply -f k8s/
          kubectl rollout restart deployment/$\{{ github.repository }}

      - name: Verify deployment
        run: |
          kubectl rollout status deployment/$\{{ github.repository }}
`

	// 创建 .github/workflows 目录
	workflowsDir := ".github/workflows"
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	return os.WriteFile(filepath.Join(workflowsDir, "ci-cd.yml"), []byte(workflow), 0644)
}

// generateGitLabCI 生成 GitLab CI 配置
func (cg *CICDGenerator) generateGitLabCI() error {
	gitlabCI := `image: golang:1.21

stages:
  - test
  - build
  - deploy

variables:
  GO_VERSION: "1.21"
  REGISTRY: registry.gitlab.com

before_script:
  - mkdir -p .go
  - go mod download

test:
  stage: test
  script:
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
  coverage: '/coverage.html'
  artifacts:
    paths:
      - coverage.html
    expire_in: 30 days

build:
  stage: build
  script:
    - CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/shode
  artifacts:
    paths:
      - app
    expire_in: 7 days

docker:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $REGISTRY
    - docker build -t $REGISTRY/$CI_PROJECT_NAME:$CI_COMMIT_SHA .
    - docker push $REGISTRY/$CI_PROJECT_NAME:$CI_COMMIT_SHA
  only:
    - main

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl apply -f k8s/
    - kubectl rollout restart deployment/$CI_PROJECT_NAME
  environment:
    name: production
  only:
    - main
`

	return os.WriteFile(".gitlab-ci.yml", []byte(gitlabCI), 0644)
}

// generateJenkins 生成 Jenkins Pipeline 配置
func (cg *CICDGenerator) generateJenkins() error {
	jenkinsfile := `pipeline {
    agent any

    tools {
        go 'go-1.21'
    }

    environment {
        REGISTRY = 'docker.io/myregistry'
    }

    stages {
        stage('Test') {
            steps {
                echo 'Running tests...'
                sh 'go test -v -race ./...'
            }
        }

        stage('Build') {
            steps {
                echo 'Building binary...'
                sh 'CGO_ENABLED=0 go build -a -installsuffix cgo -o app ./cmd/shode'
                archiveArtifacts artifacts: 'app', fingerprint: true
            }
        }

        stage('Docker Build') {
            when {
                branch 'main'
            }
            steps {
                script {
                    sh 'docker build -t $REGISTRY/$JOB_NAME:$BUILD_NUMBER .'
                }
            }
        }

        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                echo 'Deploying to Kubernetes...'
                sh 'kubectl apply -f k8s/'
                sh 'kubectl rollout restart deployment/$JOB_NAME'
            }
        }
    }

    post {
        always {
            echo 'Pipeline completed'
        }
        success {
            echo 'Pipeline succeeded!'
        }
        failure {
            echo 'Pipeline failed!'
        }
    }
}
`

	return os.WriteFile("Jenkinsfile", []byte(jenkinsfile), 0644)
}
