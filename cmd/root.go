package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "timesheet",
	Short: "A CLI tool to track your work hours using Git commits.",
	Long: `timesheet is a command-line interface (CLI) tool that helps you track your work hours
by creating special Git commits in a designated repository.

Use 'timesheet in' to start a work session.
Use 'timesheet out' to end the current work session.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
