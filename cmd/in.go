package cmd

import (
	"fmt"
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/internal/gitops"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"github.com/spf13/cobra"
	"os"
)

// inCmd represents the 'in' command
var inCmd = &cobra.Command{
	Use:   "in",
	Short: "Clock in for a new work session.",
	Run: func(cmd *cobra.Command, args []string) {
		colorutil.Cyan("Clocking IN\n")
		commitMessage := fmt.Sprintf("%s", constants.CheckInPrefix)

		if err := gitops.CreateTimeCommit(commitMessage); err != nil {
			colorutil.Red("Failed to clock in: %v\n", err)
			os.Exit(1)
		}
		colorutil.Green("Successfully clocked IN!\n")
	},
}

func init() {
	rootCmd.AddCommand(inCmd)
}
