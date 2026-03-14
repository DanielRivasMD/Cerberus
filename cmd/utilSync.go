/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/mattn/go-runewidth"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

type syncResult struct {
	name    string
	action  string // "push" or "pull"
	success bool
	message string // short message like "Already up to date", "Pushed 3 commits", etc.
	err     error
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runSyncMulti is the main entry point for the sync command.
func runSyncMulti(specificRepo string, push, pull, verbose bool) error {
	repos, err := collectRepos(specificRepo, verbose)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		return nil
	}

	action := "push"
	if pull {
		action = "pull"
	}

	results := make([]*syncResult, 0, len(repos))
	for _, r := range repos {
		res, err := syncRepository(r, push, pull, verbose)
		if err != nil {
			res = &syncResult{name: filepath.Base(r), action: action, err: err}
		}
		results = append(results, res)
	}

	printSyncTable(results)
	return nil
}

// syncRepository performs push or pull on a single repository.
func syncRepository(repoPath string, push, pull, verbose bool) (*syncResult, error) {
	name := filepath.Base(repoPath)
	action := "push"
	if pull {
		action = "pull"
	}
	res := &syncResult{name: name, action: action}

	runGit := func(args ...string) (string, string, error) {
		cmd := append([]string{"-C", repoPath}, args...)
		stdout, stderr, err := domovoi.CaptureExecCmd("git", cmd...)
		return strings.TrimSpace(stdout), strings.TrimSpace(stderr), err
	}

	// Check if repo is clean
	porcelain, _, err := runGit("status", "--porcelain")
	if err != nil {
		return res, horus.Wrap(err, "syncRepository", "git status failed")
	}
	if porcelain != "" {
		res.success = false
		res.message = "Skipped (uncommitted changes)"
		return res, nil
	}

	if pull {
		// git pull
		stdout, stderr, err := runGit("pull")
		if err != nil {
			// Pull failed (divergent, no upstream, etc.)
			res.success = false
			if strings.Contains(stderr, "divergent") {
				res.message = "Divergent branches"
			} else if strings.Contains(stderr, "no tracking information") {
				res.message = "No upstream"
			} else {
				lines := strings.Split(stderr, "\n")
				if len(lines) > 0 && lines[0] != "" {
					res.message = lines[0]
				} else {
					res.message = "Pull failed"
				}
			}
			return res, nil
		}
		// Success
		res.success = true
		if strings.Contains(stdout, "Already up to date") {
			res.message = "Already up to date"
		} else {
			lines := strings.Split(stdout, "\n")
			if len(lines) > 0 {
				res.message = lines[0]
			} else {
				res.message = "Pulled changes"
			}
		}
	} else if push {
		// git push
		stdout, stderr, err := runGit("push")
		if err != nil {
			res.success = false
			if strings.Contains(stderr, "divergent") {
				res.message = "Divergent branches (pull first)"
			} else if strings.Contains(stderr, "no upstream") {
				res.message = "No upstream (set with --set-upstream)"
			} else {
				lines := strings.Split(stderr, "\n")
				if len(lines) > 0 && lines[0] != "" {
					res.message = lines[0]
				} else {
					res.message = "Push failed"
				}
			}
			return res, nil
		}
		res.success = true
		if strings.Contains(stdout, "Everything up-to-date") {
			res.message = "Everything up-to-date"
		} else {
			lines := strings.Split(stdout, "\n")
			if len(lines) > 0 {
				res.message = lines[0]
			} else {
				res.message = "Pushed changes"
			}
		}
	}
	return res, nil
}

// printSyncTable prints a formatted Markdown table of sync results.
func printSyncTable(results []*syncResult) {
	// Column definitions
	cols := []struct {
		name      string
		alignLeft bool
	}{
		{"Repo", true},
		{"Action", true},
		{"Result", true},
		{"Message", true},
	}

	// Precompute display strings
	type rowData struct {
		repo    string
		action  string
		result  string
		message string
		err     error
	}
	rows := make([]rowData, len(results))
	maxWidths := make([]int, len(cols))

	updateMax := func(col int, s string) {
		if w := runewidth.StringWidth(s); w > maxWidths[col] {
			maxWidths[col] = w
		}
	}

	for i, r := range results {
		if r.err != nil {
			rows[i] = rowData{repo: r.name, err: r.err}
			updateMax(0, r.name)
			continue
		}

		// Action
		action := r.action
		// Result
		resultStr := "success"
		if !r.success {
			resultStr = "failed"
		}
		// Message
		msg := r.message
		if msg == "" {
			msg = "—"
		}

		rows[i] = rowData{
			repo:    r.name,
			action:  action,
			result:  resultStr,
			message: msg,
		}

		updateMax(0, r.name)
		updateMax(1, action)
		updateMax(2, resultStr)
		updateMax(3, msg)
	}

	// Headers
	headers := []string{"Repo", "Action", "Result", "Message"}
	for i, h := range headers {
		if w := runewidth.StringWidth(h); w > maxWidths[i] {
			maxWidths[i] = w
		}
	}

	// Print header
	fmt.Print("|")
	for i, h := range headers {
		if cols[i].alignLeft {
			fmt.Printf(" %s |", leftAligned(h, maxWidths[i]))
		} else {
			fmt.Printf(" %s |", rightAligned(h, maxWidths[i]))
		}
	}
	fmt.Println()

	// Separator
	fmt.Print("|")
	for i, w := range maxWidths {
		dashes := strings.Repeat("-", w)
		if cols[i].alignLeft {
			fmt.Printf(" %s |", leftAligned(dashes, w))
		} else {
			fmt.Printf(" %s |", rightAligned(dashes, w))
		}
	}
	fmt.Println()

	// Rows
	for _, r := range rows {
		if r.err != nil {
			fmt.Printf("%s: %s\n", r.repo, chalk.Red.Color("error: "+r.err.Error()))
			continue
		}

		fmt.Print("|")
		// Repo
		repoCell := leftAligned(r.repo, maxWidths[0])
		fmt.Printf(" %s |", repoCell)
		// Action
		actionCell := leftAligned(r.action, maxWidths[1])
		fmt.Printf(" %s |", actionCell)
		// Result (colored)
		resultCell := leftAligned(r.result, maxWidths[2])
		if r.result == "success" {
			resultCell = chalk.Green.Color(resultCell)
		} else if r.result == "failed" {
			resultCell = chalk.Red.Color(resultCell)
		}
		fmt.Printf(" %s |", resultCell)
		// Message
		msgCell := leftAligned(r.message, maxWidths[3])
		if strings.Contains(r.message, "uncommitted") || strings.Contains(r.message, "divergent") {
			msgCell = chalk.Yellow.Color(msgCell)
		}
		fmt.Printf(" %s |", msgCell)

		fmt.Println()
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
