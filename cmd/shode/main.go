package main

import (
	"fmt"
	"os"

	"gitee.com/com_818cloud/shode/cmd/shode/commands"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "shode",
		Short: "Shode - A secure shell script runtime platform",
		Long: `Shode is a secure shell script runtime platform that provides
modern development practices, safety features, and ecosystem tools
	for shell scripting.`,
		Version: "0.7.0",
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewRunCommand())
	rootCmd.AddCommand(commands.NewExecCommand())
	rootCmd.AddCommand(commands.NewReplCommand())
	rootCmd.AddCommand(commands.NewPkgCommand())
	rootCmd.AddCommand(commands.NewInitCommandEnhanced())
	rootCmd.AddCommand(commands.NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
