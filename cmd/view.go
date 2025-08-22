package cmd

import (
	"fmt"
	"github.com/benjameswoo1-droid/daka-tracker/internal/constants"
	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// viewCmd represents the 'view' command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show IN, OUT, lunch break times, and hours worked for each timesheet day",
	Run: func(cmd *cobra.Command, args []string) {
		if err := printTimesheetDailyView(); err != nil {
			colorutil.Red("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func printTimesheetDailyView() error {
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

	type dayRecord struct {
		InCommit    *object.Commit
		OutCommit   *object.Commit
		LunchStart  *object.Commit
		LunchEnd    *object.Commit
		WorkedHours time.Duration
	}
	records := make(map[string]*dayRecord) // key: YYYY-MM-DD

	// Iterate commits and populate records
	err = cIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When.In(aestLoc)
		dayStr := commitTime.Format("2006-01-02")
		msg := c.Message

		rec, ok := records[dayStr]
		if !ok {
			rec = &dayRecord{}
			records[dayStr] = rec
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

	// Calculate WorkedHours for each record
	for _, rec := range records {
		var worked time.Duration
		if rec.InCommit != nil && rec.OutCommit != nil {
			tIn := rec.InCommit.Author.When.In(aestLoc)
			tOut := rec.OutCommit.Author.When.In(aestLoc)
			worked = tOut.Sub(tIn)
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
		}
	}

	colorutil.Cyan("Timesheet IN/OUT, lunch break times, and total worked hours by day:\n")
	var days []string
	for day := range records {
		days = append(days, day)
	}
	sort.Strings(days)
	for _, day := range days {
		rec := records[day]
		fmt.Printf("Date: %s\n", day)
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
		// WORKED HOURS
		if rec.WorkedHours > 0 {
			hours := int(rec.WorkedHours.Hours())
			mins := int(rec.WorkedHours.Minutes()) % 60
			colorutil.Green("  Worked     : %dh %dm\n", hours, mins)
		} else {
			colorutil.Red("  Worked     : Not computable\n")
		}
		fmt.Println()
	}
	return nil
}
