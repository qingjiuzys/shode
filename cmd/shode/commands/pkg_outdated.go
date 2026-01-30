package commands

import (
	"fmt"
	"strings"

	pkgmgr "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

// newPkgOutdatedCommand creates the 'outdated' subcommand
func newPkgOutdatedCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Check for outdated packages",
		Long: `Outdated compares installed versions with the latest versions
that satisfy the version constraints in shode.json.

This helps you identify packages that can be updated.

Example:
  shode pkg outdated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := pkgmgr.NewPackageManager()
			packages, err := pm.CheckOutdated()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printOutdatedJSON(packages)
			}

			return printOutdatedTable(packages)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}

// printOutdatedTable prints outdated packages in a table format
func printOutdatedTable(packages []*OutdatedPackage) error {
	if len(packages) == 0 {
		fmt.Println("✓ All packages are up to date!")
		return nil
	}

	// Print header
	fmt.Println("\nPackage\t\tCurrent\t\tLatest\t\tType")
	fmt.Println(strings.Repeat("-", 80))

	// Print rows
	for _, pkg := range packages {
		pkgType := "dep"
		if pkg.IsDev {
			pkgType = "dev"
		}
		fmt.Printf("%s\t\t%s\t\t%s\t\t%s\n", pkg.Name, pkg.Current, pkg.Latest, pkgType)
	}

	fmt.Printf("\n%d package(s) can be updated\n", len(packages))
	return nil
}

// printOutdatedJSON prints outdated packages as JSON
func printOutdatedJSON(packages []*OutdatedPackage) error {
	// Simple JSON output
	fmt.Println("[")
	for i, pkg := range packages {
		pkgType := "dep"
		if pkg.IsDev {
			pkgType = "dev"
		}

		fmt.Printf("  {\"name\": \"%s\", \"current\": \"%s\", \"latest\": \"%s\", \"type\": \"%s\"}",
			pkg.Name, pkg.Current, pkg.Latest, pkgType)

		if i < len(packages)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("]")
	return nil
}
