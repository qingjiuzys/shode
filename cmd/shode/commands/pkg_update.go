package commands

import (
	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

// newPkgUpdateCommand creates the 'update' subcommand
func newPkgUpdateCommand() *cobra.Command {
	var latest bool
	var dev bool

	cmd := &cobra.Command{
		Use:   "update [package]",
		Short: "Update packages to their latest versions",
		Long: `Update checks for newer versions of packages and updates them.

If no package name is specified, all dependencies are checked.
Updates respect semantic versioning constraints in shode.json.

Examples:
  shode pkg update           # Update all packages
  shode pkg update logger    # Update specific package
  shode pkg update --latest  # Update to latest version (ignore semver)`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := pkgmgr.NewPackageManager()

			if len(args) == 0 {
				// Update all
				return pm.UpdateAll(dev)
			}

			// Update specific package
			packageName := args[0]

			return pm.UpdatePackage(packageName, latest, dev)
		},
	}

	cmd.Flags().BoolVarP(&latest, "latest", "l", false, "Update to latest version (ignore semver)")
	cmd.Flags().BoolVar(&dev, "dev", false, "Update dev dependencies")

	return cmd
}

// init registers the update command
func init() {
	// This will be called from pkg.go
}
