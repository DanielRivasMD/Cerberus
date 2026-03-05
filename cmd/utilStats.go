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
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoStats represents repository statistics.
type RepoStats struct {
	Repo      string
	Commit    int
	Age       string
	Language  string
	Lines     int
	Files     int
	Size      string
	Frequency map[string]int // month -> commit count
	Mean      int
	Q1        int
	Q2        int
	Q3        int
	Q4        int
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateRepoStats gathers statistics for the current repo.
func populateRepoStats(year int) (RepoStats, error) {
	stats := RepoStats{}

	// tokei
	tokeiOut, _, err := domovoi.CaptureExecCmd("tokei", "-C")
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to capture tokei output")
	}
	tokeiStats, language, err := populateTokei(tokeiOut)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to parse tokei output")
	}
	stats.Language = language + " " + strconv.Itoa(tokeiStats.Lines.Percentage) + "%"
	stats.Lines = tokeiStats.Lines.Number

	// age
	age, err := repoAge()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository age")
	}
	stats.Age = age

	// commit count
	commitCount, err := countCommits()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to count commits")
	}
	stats.Commit = commitCount

	// size
	size, err := repoSize()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository size")
	}
	stats.Size = size

	// commit frequency for given year
	freq, err := commitFrequency(year)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to fetch commit frequency")
	}
	stats.Frequency = freq

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateStatsMD generates Markdown table for stats.
func generateStatsMD(repoNames []string, year int) string {
	fieldSizes := []int{
		Defaults.repoLen,
		Defaults.commitLen,
		Defaults.ageLen,
		Defaults.languageLen,
		Defaults.linesLen,
		Defaults.sizeLen,
		Defaults.meanLen,
		Defaults.qLen,
		Defaults.qLen,
		Defaults.qLen,
		Defaults.qLen,
	}
	skip := map[string]bool{
		"Remote":    true,
		"Files":     true,
		"Frequency": true,
	}
	var sample RepoStats
	populateFunc := func(repoName string) (*RepoStats, error) {
		s, err := populateRepoStats(year)
		if err != nil {
			return nil, err
		}
		s.Repo = repoName
		return &s, nil
	}
	aligners := map[string]Alignment{
		"Repo":     AlignLeft,
		"Language": AlignLeft,
	}
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, year, aligners, RootFlags.output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// repoAge returns "Xy Ym" string.
func repoAge() (string, error) {
	out, _, err := domovoi.CaptureExecCmd("git", "log", "--reverse", "--format=%ci")
	if err != nil {
		return "", err
	}
	commitDates := strings.Split(string(out), "\n")
	if len(commitDates) == 0 || strings.TrimSpace(commitDates[0]) == "" {
		return "0y 0m", nil
	}
	firstDateStr := strings.TrimSpace(commitDates[0])
	layout := "2006-01-02 15:04:05 -0700"
	firstDate, err := time.Parse(layout, firstDateStr)
	if err != nil {
		return "0y 0m", nil
	}
	now := time.Now()
	years := now.Year() - firstDate.Year()
	months := int(now.Month()) - int(firstDate.Month())
	if months < 0 {
		years--
		months += 12
	}
	return fmt.Sprintf("%dy %dm", years, months), nil
}

// commitFrequency returns month -> commit count for given year.
func commitFrequency(year int) (map[string]int, error) {
	freq := make(map[string]int)
	for m := 1; m <= 12; m++ {
		monthKey := fmt.Sprintf("%d-%02d", year, m)
		freq[monthKey] = 0
	}

	out, _, err := domovoi.CaptureExecCmd("git", "log",
		"--since", fmt.Sprintf("%d-01-01", year),
		"--until", fmt.Sprintf("%d-12-31", year),
		"--format=%ci")
	if err != nil {
		return freq, err
	}
	commitDates := strings.Split(string(out), "\n")
	layout := "2006-01-02 15:04:05 -0700"
	for _, dateStr := range commitDates {
		if strings.TrimSpace(dateStr) == "" {
			continue
		}
		t, err := time.Parse(layout, dateStr)
		if err != nil {
			continue
		}
		monthKey := t.Format("2006-01")
		freq[monthKey]++
	}
	return freq, nil
}

// countCommits returns total number of commits.
func countCommits() (int, error) {
	out, _, err := domovoi.CaptureExecCmd("git", "rev-list", "--count", "HEAD")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// repoSize calculates total size of all files in the repo.
func repoSize() (string, error) {
	var size int64
	err := filepath.Walk(".", func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		return "", err
	}
	return formatRepoSize(int(size)), nil
}

func formatRepoSize(size int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%d GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%d MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%d KB", size/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// calculatePercentage helper (used in tokei)
func calculatePercentage(dominant, total int) int {
	if total == 0 {
		return 0
	}
	return (dominant * 100) / total
}

////////////////////////////////////////////////////////////////////////////////////////////////////
