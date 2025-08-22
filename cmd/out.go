package cmd

import (
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/internal/gitops"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

// outCmd represents the 'out' command
var outCmd = &cobra.Command{
	Use:   "out",
	Short: "Clock out of the current work session.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		colorutil.Cyan("Clocking OUT...\n")
		commitMessage := constants.CheckOutPrefix

		if err := gitops.CreateTimeCommit(commitMessage); err != nil {
			colorutil.Red("Failed to clock out: %v\n", err)
			os.Exit(1)
		}
		color.Green("Successfully clocked OUT!\n")
	},
}

// init initializes the root command and adds subcommands
func init() {
	rootCmd.AddCommand(outCmd)
}
