package lint

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/formatter"
)

// Issue represents a lint finding.
type Issue struct {
	File     string
	Line     int
	Severity string
	Message  string
}

// LintPath walks all files under the provided paths and reports issues.
func LintPath(paths []string) ([]Issue, error) {
	var issues []Issue
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return issues, err
		}

		if info.IsDir() {
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if shouldSkipDir(d.Name()) {
						return filepath.SkipDir
					}
					return nil
				}
				fileIssues, err := LintFile(path)
				if err != nil {
					return err
				}
				issues = append(issues, fileIssues...)
				return nil
			})
			if err != nil {
				return issues, err
			}
			continue
		}

		fileIssues, err := LintFile(p)
		if err != nil {
			return issues, err
		}
		issues = append(issues, fileIssues...)
	}
	return issues, nil
}

// LintFile runs heuristics against a single file.
func LintFile(path string) ([]Issue, error) {
	if !formatter.IsSupportedFile(path) {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []Issue
	scanner := bufio.NewScanner(file)
	lineNum := 0
	hasShebang := false
	hasSetErrExit := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if lineNum == 1 && strings.HasPrefix(trimmed, "#!") {
			hasShebang = true
		}
		if strings.Contains(trimmed, "set -e") || strings.Contains(trimmed, "set -o errexit") {
			hasSetErrExit = true
		}
		if strings.Contains(line, "rm -rf /") {
			issues = append(issues, Issue{
				File:     path,
				Line:     lineNum,
				Severity: "error",
				Message:  "Dangerous deletion detected (rm -rf /)",
			})
		}
		if strings.Contains(line, "TODO") {
			issues = append(issues, Issue{
				File:     path,
				Line:     lineNum,
				Severity: "info",
				Message:  "TODO left in script",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return issues, fmt.Errorf("read file: %w", err)
	}

	if !hasShebang {
		issues = append(issues, Issue{
			File:     path,
			Line:     1,
			Severity: "warning",
			Message:  "Missing shebang (#!/bin/sh)",
		})
	}
	if !hasSetErrExit {
		issues = append(issues, Issue{
			File:     path,
			Line:     0,
			Severity: "warning",
			Message:  "Consider adding 'set -e' to fail fast",
		})
	}

	return issues, nil
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "sh_models":
		return true
	}
	return strings.HasPrefix(name, ".")
}
