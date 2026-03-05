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
	"bufio"
	"os"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoDescribe represents repository features.
type RepoDescribe struct {
	Repo     string
	Overview string
	License  string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateRepoDescribe gathers info about the current repo.
func populateRepoDescribe() (RepoDescribe, error) {
	describe := RepoDescribe{}

	pwd, err := os.Getwd()
	if err != nil {
		return describe, horus.Wrap(err, "populateRepoDescribe", "failed to get current working directory")
	}
	files, err := domovoi.ListFiles(pwd)
	if err != nil {
		return describe, horus.Wrap(err, "populateRepoDescribe", "failed to list files in directory")
	}

	for _, file := range files {
		if file == "README.md" {
			desc, err := parseReadme(file, Defaults.overviewLen)
			if err != nil {
				return describe, horus.Wrap(err, "populateRepoDescribe", "failed to parse README.md")
			}
			describe.Overview = desc
		}
		if file == "LICENSE" {
			lic, err := detectLicense(file)
			if err != nil {
				return describe, horus.Wrap(err, "populateRepoDescribe", "failed to detect LICENSE")
			}
			describe.License = lic
		}
	}
	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateDescribeMD generates Markdown table for describe.
func generateDescribeMD(repoNames []string) string {
	fieldSizes := []int{Defaults.repoLen, Defaults.overviewLen, Defaults.licenseLen}
	skip := map[string]bool{"Remote": true}
	var sample RepoDescribe
	populateFunc := func(repoName string) (*RepoDescribe, error) {
		d, err := populateRepoDescribe()
		if err != nil {
			return nil, err
		}
		d.Repo = repoName
		return &d, nil
	}
	aligners := map[string]Alignment{
		"Repo":     AlignLeft,
		"Overview": AlignLeft,
		"License":  AlignLeft,
	}
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, 0, aligners, RootFlags.output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// parseReadme extracts the first paragraph under "## Overview".
func parseReadme(filename string, maxChars int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", horus.NewCategorizedHerror("parseReadme", "file_open_error", "failed to open file", err,
			map[string]any{"filename": filename})
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inDesc := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## Overview") {
			inDesc = true
			continue
		}
		if inDesc && strings.HasPrefix(line, "## ") {
			break
		}
		if inDesc {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", horus.Wrap(err, "parseReadme", "scanner error")
	}
	result := strings.Join(lines, "\n")
	result = trimmer(result)
	if len(result) > maxChars {
		result = result[:maxChars]
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// detectLicense identifies license type from file content.
func detectLicense(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", horus.Wrap(err, "detectLicense", "failed to read license file")
	}
	content := strings.ToLower(string(data))
	keywords := map[string]string{
		"mit license":                "MIT",
		"apache license":             "Apache-2.0",
		"gnu general public license": "GPL",
		"bsd license":                "BSD",
		"mozilla public license":     "MPL",
		"creative commons":           "CC",
		"eclipse public license":     "EPL",
	}
	for keyword, lic := range keywords {
		if strings.Contains(content, keyword) {
			return lic, nil
		}
	}
	return "Unknown", nil
}

// trimmer returns substring up to first period or newline.
func trimmer(desc string) string {
	if idx := strings.IndexAny(desc, ".\n"); idx >= 0 {
		return strings.TrimSpace(desc[:idx+1])
	}
	return strings.TrimSpace(desc)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
