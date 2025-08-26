package cmd

import (
	"fmt"
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/internal/gitops"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"github.com/spf13/cobra"
	"os"
)

var sickCmd = &cobra.Command{
	Use:   "sick",
	Short: "Log a sick leave for today.",
	Run: func(cmd *cobra.Command, args []string) {
		colorutil.Cyan("Logging sick leave\n")
		commitMessage := fmt.Sprintf("%s", constants.SickLeavePrefix)

		if err := gitops.CreateTimeCommit(commitMessage); err != nil {
			colorutil.Red("Failed to log sick leave: %v\n", err)
			os.Exit(1)
		}
		colorutil.Green("Successfully logged sick leave!\n")
	},
}

func init() {
	rootCmd.AddCommand(sickCmd)
}
