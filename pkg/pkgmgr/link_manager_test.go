package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLinkManager(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	lm := NewLinkManager(tmpDir)

	if lm == nil {
		t.Fatal("LinkManager 不应为 nil")
	}

	if lm.linksFile != filepath.Join(tmpDir, "shode-links.json") {
		t.Errorf("linksFile 路径不正确: got %s", lm.linksFile)
	}
}

func TestLinkManager_Link(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建测试包目录
	pkgDir := t.TempDir()
	pkgJsonPath := filepath.Join(pkgDir, "package.json")

	// 创建 package.json
	pkgJsonContent := `{"name": "@test/package", "version": "1.0.0"}`
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("创建 package.json 失败: %v", err)
	}

	// 测试链接
	err := lm.Link("@test/package", pkgDir)
	if err != nil {
		t.Fatalf("链接失败: %v", err)
	}

	// 验证链接已创建
	if !lm.IsLinked("@test/package") {
		t.Error("包应该被链接")
	}

	// 验证链接路径
	linkPath, exists := lm.GetLink("@test/package")
	if !exists {
		t.Error("链接应该存在")
	}
	if linkPath != pkgDir {
		t.Errorf("链接路径不正确: got %s, want %s", linkPath, pkgDir)
	}
}

func TestLinkManager_Link_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 测试不存在的路径
	err := lm.Link("@test/package", "/nonexistent/path")
	if err == nil {
		t.Error("应该返回错误：路径不存在")
	}
}

func TestLinkManager_Link_NoPackageJson(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建没有 package.json 的目录
	pkgDir := t.TempDir()

	err := lm.Link("@test/package", pkgDir)
	if err == nil {
		t.Error("应该返回错误：缺少 package.json")
	}
}

func TestLinkManager_Link_NameMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建测试包目录
	pkgDir := t.TempDir()
	pkgJsonPath := filepath.Join(pkgDir, "package.json")

	// 创建不同名称的 package.json
	pkgJsonContent := `{"name": "@test/other", "version": "1.0.0"}`
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("创建 package.json 失败: %v", err)
	}

	// 测试链接（包名不匹配）
	err := lm.Link("@test/package", pkgDir)
	if err == nil {
		t.Error("应该返回错误：包名不匹配")
	}
}

func TestLinkManager_Unlink(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建并链接包
	pkgDir := t.TempDir()
	pkgJsonPath := filepath.Join(pkgDir, "package.json")
	pkgJsonContent := `{"name": "@test/package", "version": "1.0.0"}`
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("创建 package.json 失败: %v", err)
	}

	if err := lm.Link("@test/package", pkgDir); err != nil {
		t.Fatalf("链接失败: %v", err)
	}

	// 取消链接
	err := lm.Unlink("@test/package")
	if err != nil {
		t.Fatalf("取消链接失败: %v", err)
	}

	// 验证链接已移除
	if lm.IsLinked("@test/package") {
		t.Error("包应该被取消链接")
	}
}

func TestLinkManager_Unlink_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	err := lm.Unlink("@test/nonexistent")
	if err == nil {
		t.Error("应该返回错误：链接不存在")
	}
}

func TestLinkManager_ListLinks(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建多个测试包
	packages := make(map[string]string)
	for i := 0; i < 3; i++ {
		pkgDir := t.TempDir()
		pkgName := filepath.Join("@test", fmt.Sprintf("package%d", i))
		packages[pkgName] = pkgDir

		pkgJsonPath := filepath.Join(pkgDir, "package.json")
		pkgJsonContent := fmt.Sprintf(`{"name": "%s", "version": "1.0.0"}`, pkgName)
		if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
			t.Fatalf("创建 package.json 失败: %v", err)
		}

		if err := lm.Link(pkgName, pkgDir); err != nil {
			t.Fatalf("链接失败: %v", err)
		}
	}

	// 列出所有链接
	links := lm.ListLinks()
	if len(links) != 3 {
		t.Errorf("应该有 3 个链接，got %d", len(links))
	}

	// 验证所有包都在列表中
	linkMap := make(map[string]string)
	for _, link := range links {
		linkMap[link.PackageName] = link.LocalPath
	}

	for pkgName, expectedPath := range packages {
		if path, exists := linkMap[pkgName]; !exists {
			t.Errorf("包 %s 不在链接列表中", pkgName)
		} else if path != expectedPath {
			t.Errorf("包 %s 路径不正确: got %s, want %s", pkgName, path, expectedPath)
		}
	}
}

func TestLinkManager_ResolveLink(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)
	modulesPath := filepath.Join(tmpDir, "sh_modules")

	// 创建并链接包
	pkgDir := t.TempDir()
	pkgJsonPath := filepath.Join(pkgDir, "package.json")
	pkgJsonContent := `{"name": "@test/package", "version": "1.0.0"}`
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("创建 package.json 失败: %v", err)
	}

	if err := lm.Link("@test/package", pkgDir); err != nil {
		t.Fatalf("链接失败: %v", err)
	}

	// 测试已链接的包
	resolvedPath := lm.ResolveLink("@test/package", modulesPath)
	if resolvedPath != pkgDir {
		t.Errorf("解析路径应该是链接路径: got %s, want %s", resolvedPath, pkgDir)
	}

	// 测试未链接的包（应该返回默认路径）
	defaultPath := lm.ResolveLink("@test/other", modulesPath)
	expectedDefault := filepath.Join(modulesPath, "@test/other")
	if defaultPath != expectedDefault {
		t.Errorf("未链接的包应该返回默认路径: got %s, want %s", defaultPath, expectedDefault)
	}
}

func TestLinkManager_Load(t *testing.T) {
	tmpDir := t.TempDir()
	lm := NewLinkManager(tmpDir)

	// 创建测试包并链接
	pkgDir := t.TempDir()
	pkgJsonPath := filepath.Join(pkgDir, "package.json")
	pkgJsonContent := `{"name": "@test/package", "version": "1.0.0"}`
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("创建 package.json 失败: %v", err)
	}

	if err := lm.Link("@test/package", pkgDir); err != nil {
		t.Fatalf("链接失败: %v", err)
	}

	// 创建新的 LinkManager 实例（测试加载）
	lm2 := NewLinkManager(tmpDir)

	if !lm2.IsLinked("@test/package") {
		t.Error("应该从文件加载链接")
	}

	linkPath, exists := lm2.GetLink("@test/package")
	if !exists || linkPath != pkgDir {
		t.Errorf("加载的链接路径不正确: got %s, want %s", linkPath, pkgDir)
	}
}
