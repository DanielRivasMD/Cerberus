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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/mattn/go-runewidth"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

type repoStatus struct {
	name     string
	path     string
	clean    bool
	ahead    int
	behind   int
	upstream string // e.g., "origin/main"
	err      error
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runStatusMulti is the main entry point for the status command.
func runStatusMulti(specificRepo string, fetch bool, verbose bool) error {
	repos, err := collectRepos(specificRepo, verbose)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		return nil
	}

	statuses := make([]*repoStatus, 0, len(repos))
	for _, r := range repos {
		stat, err := getRepoStatus(r, fetch, verbose)
		if err != nil {
			stat = &repoStatus{name: filepath.Base(r), path: r, err: err}
		}
		statuses = append(statuses, stat)
	}

	printStatusTable(statuses)
	return nil
}

// collectRepos returns list of absolute repo paths
func collectRepos(specificRepo string, verbose bool) ([]string, error) {
	if specificRepo != "" {
		abs, err := filepath.Abs(specificRepo)
		if err != nil {
			return nil, horus.Wrap(err, "collectRepos", "failed to resolve path")
		}
		ok, err := domovoi.DirExist(filepath.Join(abs, ".git"), horus.NullAction(false), verbose)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%s is not a Git repository", abs)
		}
		return []string{abs}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, horus.Wrap(err, "collectRepos", "failed to get current directory")
	}
	ok, err := domovoi.DirExist(filepath.Join(cwd, ".git"), horus.NullAction(false), verbose)
	if err != nil {
		return nil, err
	}
	if ok {
		return []string{cwd}, nil
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return nil, horus.Wrap(err, "collectRepos", "failed to read current directory")
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(cwd, e.Name())
		ok, err := domovoi.DirExist(filepath.Join(path, ".git"), horus.NullAction(false), verbose)
		if err != nil {
			continue
		}
		if ok {
			repos = append(repos, path)
		}
	}
	return repos, nil
}

// getRepoStatus gathers status for a single repository.
func getRepoStatus(repoPath string, fetch bool, verbose bool) (*repoStatus, error) {
	stat := &repoStatus{
		name: filepath.Base(repoPath),
		path: repoPath,
	}

	runGit := func(args ...string) (string, error) {
		cmd := append([]string{"-C", repoPath}, args...)
		out, _, err := domovoi.CaptureExecCmd("git", cmd...)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}

	porcelain, err := runGit("status", "--porcelain")
	if err != nil {
		return stat, horus.Wrap(err, "getRepoStatus", "git status failed")
	}
	stat.clean = (porcelain == "")

	// Fetch if requested
	if fetch {
		if _, err := runGit("fetch"); err != nil {
			stat.err = horus.Wrap(err, "getRepoStatus", "fetch failed")
		}
	}

	upstream, err := runGit("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil || upstream == "" {
		stat.upstream = "—"
		return stat, nil
	}
	stat.upstream = upstream

	out, err := runGit("rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err != nil {
		return stat, nil
	}
	parts := strings.Fields(out)
	if len(parts) == 2 {
		stat.behind, _ = strconv.Atoi(parts[0])
		stat.ahead, _ = strconv.Atoi(parts[1])
	}
	return stat, nil
}

// printStatusTable prints a formatted Markdown table with colors and proper alignment.
func printStatusTable(statuses []*repoStatus) {
	// Column definitions: name, alignment (true=left)
	cols := []struct {
		name      string
		alignLeft bool
	}{
		{"Repo", true},
		{"Clean", true},
		{"Upstream", true},
		{"Ahead", false},
		{"Behind", false},
	}

	// Precompute raw display strings for each row
	type rowData struct {
		repo     string
		clean    string
		upstream string
		ahead    string
		behind   string
		err      error
	}
	rows := make([]rowData, len(statuses))
	maxWidths := make([]int, len(cols))

	updateMax := func(col int, s string) {
		if w := runewidth.StringWidth(s); w > maxWidths[col] {
			maxWidths[col] = w
		}
	}

	for i, s := range statuses {
		if s.err != nil {
			rows[i] = rowData{repo: s.name, err: s.err}
			updateMax(0, s.name)
			continue
		}

		cleanStr := "clean"
		if !s.clean {
			cleanStr = "unclean"
		}
		upstreamStr := s.upstream
		if upstreamStr == "" {
			upstreamStr = "—"
		}
		aheadStr := strconv.Itoa(s.ahead)
		behindStr := strconv.Itoa(s.behind)

		rows[i] = rowData{
			repo:     s.name,
			clean:    cleanStr,
			upstream: upstreamStr,
			ahead:    aheadStr,
			behind:   behindStr,
		}

		updateMax(0, s.name)
		updateMax(1, cleanStr)
		updateMax(2, upstreamStr)
		updateMax(3, aheadStr)
		updateMax(4, behindStr)
	}

	// Ensure headers are at least as wide as their content
	headers := []string{"Repo", "Clean", "Upstream", "Ahead", "Behind"}
	for i, h := range headers {
		if w := runewidth.StringWidth(h); w > maxWidths[i] {
			maxWidths[i] = w
		}
	}

	// Print header row (aligned same as data)
	fmt.Print("|")
	for i, h := range headers {
		if cols[i].alignLeft {
			fmt.Printf(" %s |", leftAligned(h, maxWidths[i]))
		} else {
			fmt.Printf(" %s |", rightAligned(h, maxWidths[i]))
		}
	}
	fmt.Println()

	// Print separator row (dashes aligned)
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

	// Print data rows
	for _, r := range rows {
		if r.err != nil {
			fmt.Printf("%s: %s\n", r.repo, chalk.Red.Color("error: "+r.err.Error()))
			continue
		}

		fmt.Print("|")

		// Repo (left)
		repoCell := leftAligned(r.repo, maxWidths[0])
		fmt.Printf(" %s |", repoCell)

		// Clean (left, color after padding)
		cleanCell := leftAligned(r.clean, maxWidths[1])
		if r.clean == "unclean" {
			cleanCell = chalk.Red.Color(cleanCell)
		} else {
			cleanCell = chalk.Green.Color(cleanCell)
		}
		fmt.Printf(" %s |", cleanCell)

		// Upstream (left, dim if no upstream)
		upstreamCell := leftAligned(r.upstream, maxWidths[2])
		if r.upstream == "—" {
			upstreamCell = chalk.Dim.TextStyle(upstreamCell)
		}
		fmt.Printf(" %s |", upstreamCell)

		// Ahead (right, yellow if >0)
		aheadCell := rightAligned(r.ahead, maxWidths[3])
		if a, _ := strconv.Atoi(r.ahead); a > 0 {
			aheadCell = chalk.Yellow.Color(aheadCell)
		}
		fmt.Printf(" %s |", aheadCell)

		// Behind (right, yellow if >0)
		behindCell := rightAligned(r.behind, maxWidths[4])
		if b, _ := strconv.Atoi(r.behind); b > 0 {
			behindCell = chalk.Yellow.Color(behindCell)
		}
		fmt.Printf(" %s |", behindCell)

		fmt.Println()
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
