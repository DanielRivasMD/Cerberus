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

// collectRepos (unchanged) – returns list of absolute repo paths
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

// getRepoStatus (modified to use fetch parameter)
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
			// Non‑fatal; we still try to get ahead/behind, but mark error
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

// printStatusTable prints a formatted table with colors.
func printStatusTable(statuses []*repoStatus) {
	// Define column widths
	widths := struct{ repo, clean, upstream, ahead, behind int }{}
	for _, s := range statuses {
		if l := runewidth.StringWidth(s.name); l > widths.repo {
			widths.repo = l
		}
		cleanStr := "clean"
		if !s.clean {
			cleanStr = "unclean"
		}
		if l := runewidth.StringWidth(cleanStr); l > widths.clean {
			widths.clean = l
		}
		if l := runewidth.StringWidth(s.upstream); l > widths.upstream {
			widths.upstream = l
		}
		aheadStr := fmt.Sprintf("%d", s.ahead)
		behindStr := fmt.Sprintf("%d", s.behind)
		if l := runewidth.StringWidth(aheadStr); l > widths.ahead {
			widths.ahead = l
		}
		if l := runewidth.StringWidth(behindStr); l > widths.behind {
			widths.behind = l
		}
	}

	// Ensure minimum widths for headers
	if widths.repo < 4 {
		widths.repo = 4
	}
	if widths.clean < 5 {
		widths.clean = 5
	}
	if widths.upstream < 8 {
		widths.upstream = 8
	}
	if widths.ahead < 5 {
		widths.ahead = 5
	}
	if widths.behind < 6 {
		widths.behind = 6
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %*s  %*s\n",
		widths.repo, "Repo",
		widths.clean, "Clean",
		widths.upstream, "Upstream",
		widths.ahead, "Ahead",
		widths.behind, "Behind")
	fmt.Printf("%s  %s  %s  %s  %s\n",
		strings.Repeat("-", widths.repo),
		strings.Repeat("-", widths.clean),
		strings.Repeat("-", widths.upstream),
		strings.Repeat("-", widths.ahead),
		strings.Repeat("-", widths.behind))

	// Print rows
	for _, s := range statuses {
		if s.err != nil {
			fmt.Printf("%-*s  %s\n", widths.repo, s.name, chalk.Red.Color("error: "+s.err.Error()))
			continue
		}

		// Clean column
		cleanStr := "clean"
		cleanColor := chalk.Green
		if !s.clean {
			cleanStr = "unclean"
			cleanColor = chalk.Red
		}

		// Ahead/Behind columns
		aheadStr := fmt.Sprintf("%d", s.ahead)
		behindStr := fmt.Sprintf("%d", s.behind)
		if s.ahead > 0 {
			aheadStr = chalk.Yellow.Color(aheadStr)
		}
		if s.behind > 0 {
			behindStr = chalk.Yellow.Color(behindStr)
		}

		// Upstream column
		upstreamStr := s.upstream
		if s.upstream == "—" {
			upstreamStr = chalk.Dim.TextStyle("—")
		}

		fmt.Printf("%-*s  %-*s  %-*s  %*s  %*s\n",
			widths.repo, s.name,
			widths.clean, cleanColor.Color(cleanStr),
			widths.upstream, upstreamStr,
			widths.ahead, aheadStr,
			widths.behind, behindStr)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
