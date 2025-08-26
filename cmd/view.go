package cmd

import (
	"fmt"
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

var initialFlex float64

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show IN, OUT, lunch break times, hours worked, flex, and set initial flex for each day",
	Run: func(cmd *cobra.Command, args []string) {
		// Read INITIAL_FLEX from .env if flag not set
		if !cmd.Flags().Changed("initial-flex") {
			envFlex := os.Getenv("INITIAL_FLEX")
			if envFlex != "" {
				val, err := strconv.ParseFloat(envFlex, 64)
				if err == nil {
					initialFlex = val
				}
			}
		}
		if err := printTimesheetDailyView(initialFlex); err != nil {
			colorutil.Red("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	viewCmd.Flags().Float64Var(&initialFlex, "initial-flex", 0, "Set initial flex in hours (default 0, or from .env INITIAL_FLEX)")
	rootCmd.AddCommand(viewCmd)
}

func printTimesheetDailyView(initialFlex float64) error {
	repoPath := os.Getenv("TIMESHEET_REPO_PATH")
	if repoPath == "" {
		return fmt.Errorf("TIMESHEET_REPO_PATH environment variable is not set")
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %v", repoPath, err)
	}

	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("failed to get commit logs: %v", err)
	}

	aestLoc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		aestLoc = time.FixedZone("AEST", 10*60*60)
	}

	const standardDayHours = 7.5

	type dayRecord struct {
		InCommit       *object.Commit
		OutCommit      *object.Commit
		LunchStart     *object.Commit
		LunchEnd       *object.Commit
		WorkedHours    time.Duration
		Flex           time.Duration
		CumulativeFlex time.Duration
		IsSickLeave    bool
	}

	records := make(map[string]*dayRecord) // key: YYYY-MM-DD

	// Collect commits and populate per-day record
	err = cIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When.In(aestLoc)
		dayStr := commitTime.Format("2006-01-02")
		msg := c.Message

		rec, ok := records[dayStr]
		if !ok {
			rec = &dayRecord{}
			records[dayStr] = rec
		}
		if strings.HasPrefix(msg, constants.SickLeavePrefix) {
			rec.IsSickLeave = true
		}
		if strings.HasPrefix(msg, constants.CheckInPrefix) {
			if rec.InCommit == nil || commitTime.Before(rec.InCommit.Author.When.In(aestLoc)) {
				rec.InCommit = c
			}
		}
		if strings.HasPrefix(msg, constants.CheckOutPrefix) {
			if rec.OutCommit == nil || commitTime.After(rec.OutCommit.Author.When.In(aestLoc)) {
				rec.OutCommit = c
			}
		}
		if strings.HasPrefix(msg, constants.LunchStartPrefix) {
			if rec.LunchStart == nil || commitTime.Before(rec.LunchStart.Author.When.In(aestLoc)) {
				rec.LunchStart = c
			}
		}
		if strings.HasPrefix(msg, constants.LunchEndPrefix) {
			if rec.LunchEnd == nil || commitTime.After(rec.LunchEnd.Author.When.In(aestLoc)) {
				rec.LunchEnd = c
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(records) == 0 {
		colorutil.Red("No timesheet commits found.\n")
		return nil
	}

	// Get sorted list of days for ordered output
	var days []string
	for day := range records {
		days = append(days, day)
	}
	sort.Strings(days)

	// Calculate worked and flex hours
	cumulative := time.Duration(initialFlex * float64(time.Hour))
	for _, day := range days {
		rec := records[day]
		// If sick leave, skip hours calculation
		if rec.IsSickLeave {
			rec.WorkedHours = 0
			rec.Flex = 0
			rec.CumulativeFlex = cumulative
			continue
		}
		// Calculate worked hours
		if rec.InCommit != nil && rec.OutCommit != nil {
			tIn := rec.InCommit.Author.When.In(aestLoc)
			tOut := rec.OutCommit.Author.When.In(aestLoc)
			worked := tOut.Sub(tIn)
			// Subtract lunch break if present and valid
			if rec.LunchStart != nil && rec.LunchEnd != nil {
				tLunchStart := rec.LunchStart.Author.When.In(aestLoc)
				tLunchEnd := rec.LunchEnd.Author.When.In(aestLoc)
				lunch := tLunchEnd.Sub(tLunchStart)
				if lunch > 0 && tLunchStart.After(tIn) && tLunchEnd.Before(tOut) {
					worked -= lunch
				}
			}
			rec.WorkedHours = worked
			rec.Flex = worked - time.Duration(standardDayHours*float64(time.Hour))
		}
		// Calculate cumulative flex, starting from initial flex
		rec.CumulativeFlex = cumulative + rec.Flex
		cumulative = rec.CumulativeFlex
	}

	// Print report
	colorutil.Cyan("Timesheet IN/OUT, lunch breaks, worked hours, flex, cumulative flex by day:\n")
	fmt.Printf("Initial Flex: %.2f hours\n\n", initialFlex)
	for _, day := range days {
		rec := records[day]
		fmt.Printf("Date: %s\n", day)
		// Sick Leave display
		if rec.IsSickLeave {
			colorutil.Cyan("  SICK LEAVE: Logged for this day\n")
			fmt.Println()
			continue
		}
		// IN
		if rec.InCommit != nil {
			inTime := rec.InCommit.Author.When.In(aestLoc)
			colorutil.Green("  IN         : %s (%s)\n", inTime.Format("15:04:05 MST"), strings.TrimSpace(rec.InCommit.Message))
		} else {
			colorutil.Red("  IN         : Not found\n")
		}
		// OUT
		if rec.OutCommit != nil {
			outTime := rec.OutCommit.Author.When.In(aestLoc)
			colorutil.Green("  OUT        : %s (%s)\n", outTime.Format("15:04:05 MST"), strings.TrimSpace(rec.OutCommit.Message))
		} else {
			colorutil.Red("  OUT        : Not found\n")
		}
		// LUNCH START
		if rec.LunchStart != nil {
			lunchStartTime := rec.LunchStart.Author.When.In(aestLoc)
			colorutil.Cyan("  Lunch Start: %s (%s)\n", lunchStartTime.Format("15:04:05 MST"), strings.TrimSpace(rec.LunchStart.Message))
		} else {
			colorutil.Red("  Lunch Start: Not found\n")
		}
		// LUNCH END
		if rec.LunchEnd != nil {
			lunchEndTime := rec.LunchEnd.Author.When.In(aestLoc)
			colorutil.Cyan("  Lunch End  : %s (%s)\n", lunchEndTime.Format("15:04:05 MST"), strings.TrimSpace(rec.LunchEnd.Message))
		} else {
			colorutil.Red("  Lunch End  : Not found\n")
		}
		// WORKED HOURS, FLEX, CUMULATIVE FLEX
		if rec.WorkedHours > 0 {
			hours := int(rec.WorkedHours.Hours())
			mins := int(rec.WorkedHours.Minutes()) % 60
			colorutil.Green("  Worked     : %dh %dm\n", hours, mins)

			flexHours := int(rec.Flex.Hours())
			flexMins := int(rec.Flex.Minutes()) % 60
			if rec.Flex < 0 {
				colorutil.Red("  Flex       : %dh %dm\n", flexHours, flexMins)
			} else {
				colorutil.Green("  Flex       : +%dh %dm\n", flexHours, flexMins)
			}

			cumHours := int(rec.CumulativeFlex.Hours())
			cumMins := int(rec.CumulativeFlex.Minutes()) % 60
			colorutil.Cyan("  Cum. Flex  : %+dh %dm\n", cumHours, cumMins)
		} else {
			colorutil.Red("  Worked     : Not computable\n")
		}
		fmt.Println()
	}
	return nil
}
