package commands

import (
	"fmt"
	"os"
	"path/filepath"

	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

func newPkgLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [package] [path]",
		Short: "链接本地包进行开发",
		Long: `链接本地包到项目，用于开发和测试本地包。

用法:
  shode pkg link <package> <path>    链接本地包
  shode pkg link unlink <package>    取消链接
  shode pkg link list                列出所有链接

示例:
  shode pkg link @my/logger ./my-logger
  shode pkg link unlink @my/logger
  shode pkg link list
`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			// 获取项目根目录
			projectRoot, err := findProjectRoot()
			if err != nil {
				return fmt.Errorf("找不到项目根目录: %w", err)
			}

			linkManager := pkgmgr.NewLinkManager(projectRoot)

			// 处理子命令
			switch args[0] {
			case "unlink":
				if len(args) < 2 {
					return fmt.Errorf("请指定要取消链接的包名")
				}
				return unlinkPackage(linkManager, args[1])

			case "list":
				return listLinks(linkManager)

			default:
				// 默认是链接操作
				if len(args) < 2 {
					return fmt.Errorf("请提供包名和路径")
				}
				return linkPackage(linkManager, args[0], args[1])
			}
		},
	}

	return cmd
}

func linkPackage(lm *pkgmgr.LinkManager, packageName, localPath string) error {
	// 转换为绝对路径
	if !filepath.IsAbs(localPath) {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return fmt.Errorf("无法解析绝对路径: %w", err)
		}
		localPath = absPath
	}

	fmt.Printf("📦 链接包: %s\n", packageName)
	fmt.Printf("   路径: %s\n", localPath)

	if err := lm.Link(packageName, localPath); err != nil {
		return fmt.Errorf("链接失败: %w", err)
	}

	fmt.Println("✅ 链接成功")
	fmt.Println("\n提示: 包现在将从本地路径加载。")

	return nil
}

func unlinkPackage(lm *pkgmgr.LinkManager, packageName string) error {
	fmt.Printf("🔗 取消链接: %s\n", packageName)

	if err := lm.Unlink(packageName); err != nil {
		return fmt.Errorf("取消链接失败: %w", err)
	}

	fmt.Println("✅ 链接已移除")
	fmt.Println("\n提示: 包将从 sh_modules 目录加载。")

	return nil
}

func listLinks(lm *pkgmgr.LinkManager) error {
	links := lm.ListLinks()

	if len(links) == 0 {
		fmt.Println("📦 没有链接的包")
		return nil
	}

	fmt.Println("📦 当前链接的包:")
	fmt.Println()

	// 打印表头
	fmt.Printf("%-30s %s\n", "包名", "路径")
	fmt.Println("────────────────────────────── ──────────────────────────────────────")

	// 打印每个链接
	for _, link := range links {
		fmt.Printf("%-30s %s\n", link.PackageName, link.LocalPath)
	}

	fmt.Printf("\n共 %d 个链接\n", len(links))

	return nil
}

// findProjectRoot 查找项目根目录（包含 shode.json 的目录）
func findProjectRoot() (string, error) {
 cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		// 检查是否存在 shode.json
		if _, err := os.Stat(filepath.Join(dir, "shode.json")); err == nil {
			return dir, nil
		}

		// 到达根目录
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 shode.json，请确保在 Shode 项目目录中运行此命令")
		}

		dir = parent
	}
}
