////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func repoAge() (string, error) {
	// get first commit
	out, _, ε := domovoi.CaptureExecCmd("git", "log", "--reverse", "--format=%ci")
	horus.CheckErr(ε)

	// split output into individual lines
	commitDates := strings.Split(string(out), "\n")
	if len(commitDates) == 0 || strings.TrimSpace(commitDates[0]) == "" {
		return "", fmt.Errorf("no commit dates found in the repository")
	}

	// use oldest commit date
	firstCommitDateStr := strings.TrimSpace(commitDates[0])

	// parse first commit date
	layout := "2006-01-02 15:04:05 -0700" // git commit date format
	firstCommitDate, ε := time.Parse(layout, firstCommitDateStr)
	horus.CheckErr(ε)

	// calculate difference
	currentDate := time.Now()
	years := currentDate.Year() - firstCommitDate.Year()
	months := int(currentDate.Month()) - int(firstCommitDate.Month())

	// Adjust for negative months (e.g., December - February)
	if months < 0 {
		years--
		months += 12
	}

	// Format the result
	formattedAge := fmt.Sprintf("%dy %dm", years, months)

	return formattedAge, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
