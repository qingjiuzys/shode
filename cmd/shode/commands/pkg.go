package commands

import (
	"fmt"
	"strings"

	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

// NewPkgCommand creates the 'pkg' command for package management
func NewPkgCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Manage Shode package dependencies",
		Long: `Package management commands for handling dependencies and scripts
in Shode projects. Uses shode.json for configuration.`,
	}

	// Add subcommands
	cmd.AddCommand(newPkgInitCommand())
	cmd.AddCommand(newPkgInstallCommand())
	cmd.AddCommand(newPkgAddCommand())
	cmd.AddCommand(newPkgRemoveCommand())
	cmd.AddCommand(newPkgListCommand())
	cmd.AddCommand(newPkgRunCommand())
	cmd.AddCommand(newPkgScriptCommand())
	cmd.AddCommand(newPkgSearchCommand())
	cmd.AddCommand(newPkgPublishCommand())
	cmd.AddCommand(newPkgSignerCommand())

	return cmd
}

// newPkgInitCommand creates the 'init' subcommand
func newPkgInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [name] [version]",
		Short: "Initialize a new package configuration",
		Long:  `Init creates a new shode.json file with basic package configuration.`,
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "my-shode-project"
			version := "1.0.0"

			if len(args) > 0 {
				name = args[0]
			}
			if len(args) > 1 {
				version = args[1]
			}

			pm := pkgmgr.NewPackageManager()
			if err := pm.Init(name, version); err != nil {
				return fmt.Errorf("failed to initialize package: %v", err)
			}

			fmt.Printf("Initialized package %s@%s\n", name, version)
			fmt.Println("shode.json created successfully!")
			return nil
		},
	}
}

// newPkgInstallCommand creates the 'install' subcommand
func newPkgInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install all dependencies",
		Long:  `Install downloads and installs all dependencies specified in shode.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			allowUnsigned, _ := cmd.Flags().GetBool("allow-unsigned")
			pm := pkgmgr.NewPackageManager()
			return pm.Install(&pkgmgr.InstallOptions{AllowUnsigned: allowUnsigned})
		},
	}

	cmd.Flags().Bool("allow-unsigned", false, "Allow installing unsigned packages")
	return cmd
}

// newPkgAddCommand creates the 'add' subcommand
func newPkgAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [package] [version]",
		Short: "Add a package dependency",
		Long:  `Add installs a package and adds it to the dependencies.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]
			version := "latest"
			if len(args) > 1 {
				version = args[1]
			}

			// Check if it's a dev dependency
			dev, _ := cmd.Flags().GetBool("dev")

			pm := pkgmgr.NewPackageManager()
			if err := pm.AddDependency(packageName, version, dev); err != nil {
				return fmt.Errorf("failed to add dependency: %v", err)
			}

			depType := "dependency"
			if dev {
				depType = "dev dependency"
			}
			fmt.Printf("Added %s %s@%s\n", depType, packageName, version)
			return nil
		},
	}

	cmd.Flags().BoolP("dev", "D", false, "Add as dev dependency")
	return cmd
}

// newPkgRemoveCommand creates the 'remove' subcommand
func newPkgRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [package]",
		Short: "Remove a package dependency",
		Long:  `Remove uninstalls a package and removes it from the dependencies.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]
			dev, _ := cmd.Flags().GetBool("dev")

			pm := pkgmgr.NewPackageManager()
			if err := pm.RemoveDependency(packageName, dev); err != nil {
				return fmt.Errorf("failed to remove dependency: %v", err)
			}

			depType := "dependency"
			if dev {
				depType = "dev dependency"
			}
			fmt.Printf("Removed %s %s\n", depType, packageName)
			return nil
		},
	}

	cmd.Flags().BoolP("dev", "D", false, "Remove from dev dependencies")
	return cmd
}

// newPkgListCommand creates the 'list' subcommand
func newPkgListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all dependencies",
		Long:  `List displays all dependencies from shode.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := pkgmgr.NewPackageManager()
			return pm.ListDependencies()
		},
	}
}

// newPkgRunCommand creates the 'run' subcommand
func newPkgRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run [script]",
		Short: "Run a package script",
		Long:  `Run executes a script defined in the scripts section of shode.json.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptName := args[0]

			pm := pkgmgr.NewPackageManager()
			return pm.RunScript(scriptName)
		},
	}
}

// newPkgScriptCommand creates the 'script' subcommand
func newPkgScriptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "script [name] [command]",
		Short: "Add or manage package scripts",
		Long: `Script manages scripts in shode.json. 
Without arguments, lists all scripts. With name and command, adds a new script.`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := pkgmgr.NewPackageManager()

			if len(args) == 0 {
				// List all scripts
				if err := pm.LoadConfig(); err != nil {
					return err
				}

				config := pm.GetConfig()
				if len(config.Scripts) == 0 {
					fmt.Println("No scripts defined in shode.json")
					return nil
				}

				fmt.Println("Scripts:")
				for name, command := range config.Scripts {
					fmt.Printf("  %s: %s\n", name, command)
				}
				return nil
			}

			if len(args) == 1 {
				return fmt.Errorf("script command requires both name and command")
			}

			// Add new script
			name := args[0]
			command := strings.Join(args[1:], " ")

			if err := pm.AddScript(name, command); err != nil {
				return fmt.Errorf("failed to add script: %v", err)
			}

			fmt.Printf("Added script '%s': %s\n", name, command)
			return nil
		},
	}

	return cmd
}

// newPkgSearchCommand creates the 'search' subcommand
func newPkgSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search for packages in the registry",
		Long:  `Search queries the Shode package registry for packages matching the search term.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			
			pm := pkgmgr.NewPackageManager()
			results, err := pm.Search(query)
			if err != nil {
				return fmt.Errorf("search failed: %v", err)
			}

			if len(results) == 0 {
				fmt.Println("No packages found")
				return nil
			}

			fmt.Printf("Found %d package(s):\n\n", len(results))
			for i, result := range results {
				fmt.Printf("%d. %s@%s\n", i+1, result.Name, result.Version)
				fmt.Printf("   Description: %s\n", result.Description)
				if result.Author != "" {
					fmt.Printf("   Author: %s\n", result.Author)
				}
				if len(result.Keywords) > 0 {
					fmt.Printf("   Keywords: %v\n", result.Keywords)
				}
				fmt.Printf("   Downloads: %d", result.Downloads)
				if result.Verified {
					fmt.Printf(" ✓ Verified")
				}
				fmt.Println()
				fmt.Println()
			}

			return nil
		},
	}
}

// newPkgPublishCommand creates the 'publish' subcommand
func newPkgPublishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish the package to the registry",
		Long:  `Publish uploads the current package to the Shode package registry.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			signer, _ := cmd.Flags().GetString("signer")
			keyPath, _ := cmd.Flags().GetString("key")

			pm := pkgmgr.NewPackageManager()
			
			// Load config to verify package is ready
			if err := pm.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load package config: %v", err)
			}

			config := pm.GetConfig()
			fmt.Printf("Publishing %s@%s to registry...\n", config.Name, config.Version)

			opts := &pkgmgr.PublishOptions{
				SignerID: signer,
				KeyPath:  keyPath,
			}

			if err := pm.Publish(opts); err != nil {
				return fmt.Errorf("publish failed: %v", err)
			}

			fmt.Printf("✓ Successfully published %s@%s\n", config.Name, config.Version)
			return nil
		},
	}

	cmd.Flags().String("signer", "", "Signer ID to use for publishing")
	cmd.Flags().String("key", "", "Path to signing key (.ed25519)")
	return cmd
}
