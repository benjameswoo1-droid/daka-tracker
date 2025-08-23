package cmd

import (
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/internal/gitops"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"github.com/spf13/cobra"
	"os"
)

// Parent lunch command
var lunchCmd = &cobra.Command{
	Use:   "lunch",
	Short: "Commands related to lunch breaks",
}

// lunch in
var lunchInCmd = &cobra.Command{
	Use:   "in",
	Short: "Record the start of a lunch break.",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		message := constants.LunchStartPrefix
		if err := gitops.CreateTimeCommit(message); err != nil {
			colorutil.Red("Failed to record lunch start: %v\n", err)
			os.Exit(1)
		}
		colorutil.Green("Lunch break started!\n")
	},
}

// lunch out
var lunchOutCmd = &cobra.Command{
	Use:   "out",
	Short: "Record the end of a lunch break.",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		message := constants.LunchEndPrefix
		if err := gitops.CreateTimeCommit(message); err != nil {
			colorutil.Red("Failed to record lunch end: %v\n", err)
			os.Exit(1)
		}
		colorutil.Green("Lunch break ended!\n")
	},
}

func init() {
	lunchCmd.AddCommand(lunchInCmd)
	lunchCmd.AddCommand(lunchOutCmd)
	rootCmd.AddCommand(lunchCmd)
}
