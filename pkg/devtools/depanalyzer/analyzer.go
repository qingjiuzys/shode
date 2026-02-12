// Package depanalyzer 提供依赖分析功能
package depanalyzer

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Analyzer 依赖分析器
type Analyzer struct {
	fset        *token.FileSet
	packages    map[string]*PackageInfo
	ignoreList  map[string]bool
	workingDir  string
}

// PackageInfo 包信息
type PackageInfo struct {
	Name         string
	ImportPath   string
	Imports      []string
	Files        []string
	Dependencies []*PackageInfo
	Dependents   []*PackageInfo
	Cycle        bool
}

// DependencyInfo 依赖信息
type DependencyInfo struct {
	Internal     []string
	External     []string
	StdLib       []string
	Cycles       []string
	Unused       []string
}

// NewAnalyzer 创建分析器
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		fset:       token.NewFileSet(),
		packages:   make(map[string]*PackageInfo),
		ignoreList: make(map[string]bool),
	}
}

// IgnorePackage 添加到忽略列表
func (a *Analyzer) IgnorePackage(pkg string) {
	a.ignoreList[pkg] = true
}

// Analyze 分析目录
func (a *Analyzer) Analyze(dir string) error {
	a.workingDir = dir
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏目录和 vendor
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
		}

		// 只处理 Go 文件
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 跳过测试文件
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return a.analyzeFile(path)
	})
}

// analyzeFile 分析单个文件
func (a *Analyzer) analyzeFile(path string) error {
	file, err := parser.ParseFile(a.fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	pkgName := file.Name.Name

	// 获取包的导入路径
	importPath, err := filepath.Rel(a.workingDir, filepath.Dir(path))
	if err != nil {
		return err
	}
	importPath = filepath.ToSlash(importPath)

	// 查找或创建包信息
	pkg, ok := a.packages[importPath]
	if !ok {
		pkg = &PackageInfo{
			Name:       pkgName,
			ImportPath: importPath,
			Imports:    make([]string, 0),
			Files:      make([]string, 0),
		}
		a.packages[importPath] = pkg
	}

	// 添加文件
	pkg.Files = append(pkg.Files, path)

	// 收集导入
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if !a.shouldIgnore(importPath) {
			pkg.Imports = append(pkg.Imports, importPath)
		}
	}

	return nil
}

// shouldIgnore 检查是否应该忽略
func (a *Analyzer) shouldIgnore(importPath string) bool {
	// 检查忽略列表
	for ignore := range a.ignoreList {
		if strings.HasPrefix(importPath, ignore) {
			return true
		}
	}

	// 忽略标准库
	if !strings.Contains(importPath, ".") {
		return false
	}

	return false
}

// GetDependencies 获取依赖信息
func (a *Analyzer) GetDependencies() *DependencyInfo {
	info := &DependencyInfo{
		Internal: make([]string, 0),
		External: make([]string, 0),
		StdLib:   make([]string, 0),
		Cycles:   make([]string, 0),
		Unused:   make([]string, 0),
	}

	seen := make(map[string]bool)

	for _, pkg := range a.packages {
		for _, imp := range pkg.Imports {
			if seen[imp] {
				continue
			}
			seen[imp] = true

			if !strings.Contains(imp, ".") {
				// 标准库
				info.StdLib = append(info.StdLib, imp)
			} else if a.isInternal(imp) {
				// 内部包
				info.Internal = append(info.Internal, imp)
			} else {
				// 外部包
				info.External = append(info.External, imp)
			}
		}
	}

	// 检测循环依赖
	info.Cycles = a.detectCycles()

	// 排序
	sort.Strings(info.Internal)
	sort.Strings(info.External)
	sort.Strings(info.StdLib)
	sort.Strings(info.Cycles)

	return info
}

// isInternal 检查是否为内部包
func (a *Analyzer) isInternal(importPath string) bool {
	for pkgPath := range a.packages {
		if strings.HasPrefix(importPath, pkgPath) {
			return true
		}
	}
	return false
}

// detectCycles 检测循环依赖
func (a *Analyzer) detectCycles() []string {
	cycles := make([]string, 0)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for pkgPath := range a.packages {
		if !visited[pkgPath] {
			if cycle := a.detectCycle(pkgPath, visited, recStack); cycle != "" {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// detectCycle 检测单个循环
func (a *Analyzer) detectCycle(pkgPath string, visited, recStack map[string]bool) string {
	visited[pkgPath] = true
	recStack[pkgPath] = true

	pkg, ok := a.packages[pkgPath]
	if !ok {
		recStack[pkgPath] = false
		return ""
	}

	for _, imp := range pkg.Imports {
		if !a.isInternal(imp) {
			continue
		}

		if !visited[imp] {
			if cycle := a.detectCycle(imp, visited, recStack); cycle != "" {
				return cycle
			}
		} else if recStack[imp] {
			return fmt.Sprintf("%s -> %s", pkgPath, imp)
		}
	}

	recStack[pkgPath] = false
	return ""
}

// PrintReport 打印分析报告
func (a *Analyzer) PrintReport() {
	info := a.GetDependencies()

	fmt.Println("\n=== Dependency Analysis Report ===\n")

	fmt.Printf("Total Packages: %d\n\n", len(a.packages))

	fmt.Printf("Internal Dependencies (%d):\n", len(info.Internal))
	for _, dep := range info.Internal {
		fmt.Printf("  - %s\n", dep)
	}

	fmt.Printf("\nExternal Dependencies (%d):\n", len(info.External))
	for _, dep := range info.External {
		fmt.Printf("  - %s\n", dep)
	}

	fmt.Printf("\nStandard Library Dependencies (%d):\n", len(info.StdLib))
	for _, dep := range info.StdLib {
		fmt.Printf("  - %s\n", dep)
	}

	if len(info.Cycles) > 0 {
		fmt.Printf("\nCyclic Dependencies (%d):\n", len(info.Cycles))
		for _, cycle := range info.Cycles {
			fmt.Printf("  - %s\n", cycle)
		}
	} else {
		fmt.Println("\n✓ No cyclic dependencies detected")
	}

	fmt.Println("\n==================================\n")
}

// GetPackageStatistics 获取包统计信息
func (a *Analyzer) GetPackageStatistics() map[string]int {
	stats := make(map[string]int)

	for _, pkg := range a.packages {
		stats["total"]++
		stats["files"] += len(pkg.Files)
		stats["imports"] += len(pkg.Imports)
	}

	return stats
}

// FindUnusedPackages 查找未使用的包
func (a *Analyzer) FindUnusedPackages() []string {
	imported := make(map[string]bool)

	// 收集所有被导入的包
	for _, pkg := range a.packages {
		for _, imp := range pkg.Imports {
			if a.isInternal(imp) {
				imported[imp] = true
			}
		}
	}

	// 查找未导入的内部包
	unused := make([]string, 0)
	for pkgPath := range a.packages {
		if !imported[pkgPath] {
			unused = append(unused, pkgPath)
		}
	}

	sort.Strings(unused)
	return unused
}

// GetImportTree 获取导入树
func (a *Analyzer) GetImportTree(pkgPath string, depth int) string {
	pkg, ok := a.packages[pkgPath]
	if !ok {
		return ""
	}

	var sb strings.Builder
	indent := strings.Repeat("  ", depth)

	sb.WriteString(fmt.Sprintf("%s%s\n", indent, pkgPath))

	for _, imp := range pkg.Imports {
		if a.isInternal(imp) {
			sb.WriteString(a.GetImportTree(imp, depth+1))
		}
	}

	return sb.String()
}
