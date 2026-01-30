package commands

import (
	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
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
