package tester

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

// TestResult captures a single test outcome.
type TestResult struct {
	Name     string
	File     string
	Passed   bool
	ExitCode int
	Stdout   string
	Stderr   string
	Failures []string
	Duration time.Duration
}

// Run executes tests for all discovered files under paths.
func Run(paths []string) ([]TestResult, error) {
	files, err := discoverFiles(paths)
	if err != nil {
		return nil, err
	}

	var results []TestResult
	for _, file := range files {
		res, err := runSingleTest(file)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

func runSingleTest(path string) (TestResult, error) {
	spec, err := parseSpec(path)
	if err != nil {
		return TestResult{}, err
	}

	env := environment.NewEnvironmentManager()
	_ = env.ChangeDir(filepath.Dir(path))
	execEngine := engine.NewExecutionEngine(
		env,
		stdlib.New(),
		module.NewModuleManager(),
		sandbox.NewSecurityChecker(),
	)

	p := parser.NewSimpleParser()
	script, err := p.ParseFile(path)
	if err != nil {
		return TestResult{}, fmt.Errorf("parse %s: %w", path, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	execResult, err := execEngine.Execute(ctx, script)
	duration := time.Since(start)
	if err != nil {
		return TestResult{}, err
	}

	stdout, stderr, exitCode := summarize(execResult.Commands)
	passed, failures := evaluate(stdout, stderr, exitCode, spec)

	return TestResult{
		Name:     spec.Name,
		File:     path,
		Passed:   passed,
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Failures: failures,
		Duration: duration,
	}, nil
}

type expectations struct {
	Name           string
	ExpectContains []string
	ExitCode       int
}

func parseSpec(path string) (expectations, error) {
	file, err := os.Open(path)
	if err != nil {
		return expectations{}, err
	}
	defer file.Close()

	spec := expectations{
		Name:     filepath.Base(path),
		ExitCode: 0,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "# EXPECT:"):
			spec.ExpectContains = append(spec.ExpectContains, strings.TrimSpace(line[len("# EXPECT:"):]))
		case strings.HasPrefix(line, "# EXIT:"):
			value := strings.TrimSpace(line[len("# EXIT:"):])
			fmt.Sscanf(value, "%d", &spec.ExitCode)
		case strings.HasPrefix(line, "# NAME:"):
			spec.Name = strings.TrimSpace(line[len("# NAME:"):])
		}
	}

	return spec, scanner.Err()
}

func summarize(commands []*engine.CommandResult) (stdout string, stderr string, exitCode int) {
	var outBuilder strings.Builder
	var errBuilder strings.Builder
	exitCode = 0

	for _, cmd := range commands {
		if cmd == nil {
			continue
		}
		if cmd.Output != "" {
			outBuilder.WriteString(cmd.Output)
		}
		if cmd.Error != "" {
			errBuilder.WriteString(cmd.Error)
		}
		if !cmd.Success {
			exitCode = cmd.ExitCode
		}
	}

	return outBuilder.String(), errBuilder.String(), exitCode
}

func evaluate(stdout, stderr string, exitCode int, spec expectations) (bool, []string) {
	var failures []string

	if exitCode != spec.ExitCode {
		failures = append(failures, fmt.Sprintf("expected exit %d, got %d", spec.ExitCode, exitCode))
	}
	for _, expect := range spec.ExpectContains {
		if !strings.Contains(stdout, expect) {
			failures = append(failures, fmt.Sprintf("stdout missing %q", expect))
		}
	}

	return len(failures) == 0, failures
}

func discoverFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
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
				if isTestFile(path) {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		if isTestFile(p) {
			files = append(files, p)
		}
	}
	return files, nil
}

func isTestFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".sh" && ext != ".sho" && ext != ".shode" {
		return false
	}
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test"+ext) || strings.Contains(strings.ToLower(path), "/tests/")
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "sh_models":
		return true
	}
	return strings.HasPrefix(name, ".")
}
