package commands

import (
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/scaffold"
	"github.com/spf13/cobra"
)

// NewInitCommandEnhanced creates the enhanced 'init' command with scaffolding support
func NewInitCommandEnhanced() *cobra.Command {
	var templateType string
	var version string
	var description string
	var port string
	var listTemplates bool

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Create a new Shode project",
		Long: `Init creates a new Shode project with modern scaffolding.

Supported project types:
  - basic:       Basic Shode project with package management
  - web-service: Web service with HTTP and config packages
  - cli-tool:    CLI tool project structure

Examples:
  shode init myproject
  shode init myapp --type=web-service
  shode init mytool --type=cli-tool --version=2.0.0
  shode init --list-templates`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 列出模板
			if listTemplates {
				return listTemplatesCmd()
			}

			// 检查项目名称
			if len(args) == 0 {
				return fmt.Errorf("请提供项目名称\n\n使用 'shode init --list-templates' 查看可用模板")
			}

			projectName := args[0]

			// 验证项目名称
			if err := scaffold.ValidateProjectName(projectName); err != nil {
				return err
			}

			// 设置默认模板类型
			if templateType == "" {
				templateType = "basic"
			}

			// 准备选项
			options := make(map[string]string)
			if version != "" {
				options["version"] = version
			}
			if description != "" {
				options["description"] = description
			}
			if port != "" {
				options["port"] = port
			}

			// 创建生成器并生成项目
			gen := scaffold.NewGenerator()
			if err := gen.Generate(projectName, templateType, options); err != nil {
				return err
			}

			return nil
		},
	}

	// 添加标志
	cmd.Flags().StringVarP(&templateType, "type", "t", "basic", "项目类型 (basic, web-service, cli-tool)")
	cmd.Flags().StringVarP(&version, "version", "v", "1.0.0", "项目版本号")
	cmd.Flags().StringVarP(&description, "description", "d", "", "项目描述")
	cmd.Flags().StringVarP(&port, "port", "p", "8080", "服务端口（仅 web-service）")
	cmd.Flags().BoolVarP(&listTemplates, "list-templates", "l", false, "列出所有可用模板")

	return cmd
}

// listTemplatesCmd 列出所有可用模板
func listTemplatesCmd() error {
	gen := scaffold.NewGenerator()
	templates := gen.ListTemplates()

	fmt.Println("可用的项目模板:")
	fmt.Println()

	for _, tmpl := range templates {
		fmt.Printf("  📦 %-15s %s\n", tmpl.Name, tmpl.Description)
	}

	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  shode init <project-name> --type=<template-name>")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  shode init myapp --type=basic")
	fmt.Println("  shode init myservice --type=web-service --port=3000")
	fmt.Println("  shode init mytool --type=cli-tool")

	return nil
}
