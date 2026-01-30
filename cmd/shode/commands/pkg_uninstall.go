package commands

import (
	"fmt"

	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

// newPkgUninstallCommand creates the 'uninstall' subcommand
func newPkgUninstallCommand() *cobra.Command {
	var dev bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "uninstall <package>",
		Short: "Uninstall a package",
		Long: `Uninstall removes a package from your project.

This will:
1. Remove the package from shode.json
2. Delete the package files from sh_modules/
3. Update the lock file

Example:
  shode pkg uninstall logger
  shode pkg uninstall @shode/logger --dev`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			pm := pkgmgr.NewPackageManager()

			if dryRun {
				fmt.Printf("Would uninstall: %s (dev=%v)\n", packageName, dev)
				return nil
			}

			return pm.Uninstall(packageName, dev)
		},
	}

	cmd.Flags().BoolVarP(&dev, "dev", "D", false, "Uninstall from dev dependencies")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be uninstalled without actually uninstalling")

	return cmd
}
