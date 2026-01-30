package commands

import (
	"fmt"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

// newPkgInfoCommand creates the 'info' subcommand
func newPkgInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <package>",
		Short: "View detailed package information",
		Long: `Info displays detailed information about a package including:
- Latest version
- All available versions
- Description
- Author
- Dependencies
- Download statistics

Example:
  shode pkg info @shode/logger`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			pm := pkgmgr.NewPackageManager()
			return pm.ShowPackageInfo(packageName)
		},
	}

	return cmd
}

// formatPackageInfo formats package info for display
func formatPackageInfo(info *PackageDisplayInfo) string {
	var builder strings.Builder

	// Package name and version
	builder.WriteString(fmt.Sprintf("\n%s@%s\n", info.Name, info.LatestVersion))

	// Description
	if info.Description != "" {
		builder.WriteString(fmt.Sprintf("\n├─ Description: %s\n", info.Description))
	}

	// Metadata
	builder.WriteString("├─ Metadata\n")
	if info.Author != "" {
		builder.WriteString(fmt.Sprintf("│  ├─ Author: %s\n", info.Author))
	}
	if info.License != "" {
		builder.WriteString(fmt.Sprintf("│  ├─ License: %s\n", info.License))
	}
	if info.Homepage != "" {
		builder.WriteString(fmt.Sprintf("│  ├─ Homepage: %s\n", info.Homepage))
	}
	if info.Repository != "" {
		builder.WriteString(fmt.Sprintf("│  └─ Repository: %s\n", info.Repository))
	}

	// Versions
	if len(info.Versions) > 0 {
		builder.WriteString(fmt.Sprintf("├─ Versions: %s\n", strings.Join(info.Versions, ", ")))
	}

	// Dependencies
	if len(info.Dependencies) > 0 {
		builder.WriteString("├─ Dependencies\n")
		for dep, version := range info.Dependencies {
			builder.WriteString(fmt.Sprintf("│  └─ %s: %s\n", dep, version))
		}
	}

	// Statistics
	builder.WriteString("└─ Statistics\n")
	builder.WriteString(fmt.Sprintf("   ├─ Downloads: %d\n", info.Downloads))
	if info.Verified {
		builder.WriteString("   └─ Verified: ✓ Yes\n")
	} else {
		builder.WriteString("   └─ Verified: ✗ No\n")
	}

	// Installed version
	if info.InstalledVersion != "" {
		if info.InstalledVersion != info.LatestVersion {
			builder.WriteString(fmt.Sprintf("\n⚠️  Installed: %s (Update available: %s)\n",
				info.InstalledVersion, info.LatestVersion))
		} else {
			builder.WriteString(fmt.Sprintf("\n✓ Installed: %s (latest)\n", info.InstalledVersion))
		}
	}

	return builder.String()
}
