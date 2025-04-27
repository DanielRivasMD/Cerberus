////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func repoAge(repo string) (string, error) {
	// get first commit
	out, _, ε := execCmdCapture("git", "-C", repo, "log", "--reverse", "--format=%ci")
	checkErr(ε)

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
	checkErr(ε)

	// calculate difference
	currentDate := time.Now()
	repoAge := currentDate.Sub(firstCommitDate)

	// format
	years := int(repoAge.Hours() / (24 * 365))
	days := int(repoAge.Hours()/(24)) % 365
	formattedAge := fmt.Sprintf("%dy %dd", years, days)

	return formattedAge, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
