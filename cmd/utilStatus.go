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
	// Determine list of repositories to check
	repos, err := collectRepos(specificRepo, verbose)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		return nil
	}

	// Collect status for each repo
	statuses := make([]*repoStatus, 0, len(repos))
	for _, r := range repos {
		stat, err := getRepoStatus(r, fetch, verbose)
		if err != nil {
			// If error, still include with error field
			stat = &repoStatus{name: filepath.Base(r), path: r, err: err}
		}
		statuses = append(statuses, stat)
	}

	// Print results
	printStatuses(statuses)
	return nil
}

// collectRepos returns a list of absolute paths to Git repositories.
// If specificRepo is given, it checks that path.
// Otherwise, if current directory is a Git repo, returns that.
// Otherwise, scans immediate subdirectories for .git.
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

	// Check current directory
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

	// Scan subdirectories
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
			continue // skip problematic entries
		}
		if ok {
			repos = append(repos, path)
		}
	}
	return repos, nil
}

// getRepoStatus gathers status information for a single repository.
func getRepoStatus(repoPath string, fetch bool, verbose bool) (*repoStatus, error) {
	stat := &repoStatus{
		name: filepath.Base(repoPath),
		path: repoPath,
	}

	// Helper to run git commands inside repo
	runGit := func(args ...string) (string, error) {
		cmd := append([]string{"-C", repoPath}, args...)
		out, _, err := domovoi.CaptureExecCmd("git", cmd...)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}

	// Check if clean (no unstaged changes)
	porcelain, err := runGit("status", "--porcelain")
	if err != nil {
		return stat, horus.Wrap(err, "getRepoStatus", "git status failed")
	}
	stat.clean = (porcelain == "")

	// Optionally fetch
	if fetch {
		if _, err := runGit("fetch"); err != nil {
			// Non‑fatal; we still try to get ahead/behind, but mark error
			stat.err = horus.Wrap(err, "getRepoStatus", "fetch failed")
		}
	}

	// Get upstream branch (if any)
	upstream, err := runGit("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil || upstream == "" {
		// No upstream configured
		stat.upstream = "no upstream"
		return stat, nil
	}
	stat.upstream = upstream

	// Get ahead/behind counts
	// Format: "ahead X, behind Y" from `git rev-list --count --left-right @{upstream}...HEAD`
	out, err := runGit("rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err != nil {
		// Could be that upstream doesn't exist yet (new branch)
		return stat, nil
	}
	parts := strings.Fields(out)
	if len(parts) == 2 {
		stat.behind, _ = strconv.Atoi(parts[0]) // left side: commits behind (on upstream not in local)
		stat.ahead, _ = strconv.Atoi(parts[1])  // right side: commits ahead
	}
	return stat, nil
}

// printStatuses outputs the collected status information with colors.
func printStatuses(statuses []*repoStatus) {
	for _, s := range statuses {
		line := s.name + ": "

		if s.err != nil {
			line += chalk.Red.Color("error: " + s.err.Error())
			fmt.Println(line)
			continue
		}

		// Cleanliness
		if s.clean {
			line += chalk.Green.Color("clean")
		} else {
			line += chalk.Red.Color("unclean")
		}

		// Upstream info
		if s.upstream == "" {
			line += chalk.Yellow.Color(" (no upstream)")
		} else if s.upstream == "no upstream" {
			line += chalk.Yellow.Color(" (no upstream)")
		} else {
			line += fmt.Sprintf(" (%s", s.upstream)
			if s.ahead == 0 && s.behind == 0 {
				line += chalk.Green.Color(", up-to-date")
			} else {
				if s.ahead > 0 {
					line += fmt.Sprintf(", ahead %d", s.ahead)
				}
				if s.behind > 0 {
					line += fmt.Sprintf(", behind %d", s.behind)
				}
			}
			line += ")"
		}
		fmt.Println(line)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
