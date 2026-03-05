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
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// cloneRepository clones a Git repository into targetDir.
func cloneRepository(repoURL, targetDir string) error {
	out, _, err := domovoi.CaptureExecCmd("git", "clone", repoURL, targetDir)
	if err != nil {
		return horus.Wrap(err, "cloneRepository", "failed to clone repository: "+repoURL)
	}
	_ = strings.TrimSpace(string(out))
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// cloneRepositoriesFromCSV reads a CSV file and clones each repo.
func cloneRepositoriesFromCSV(csvFile, targetDir string) error {
	if strings.TrimSpace(targetDir) == "" {
		targetDir = "."
	}

	// Verify CSV exists
	_, err := domovoi.FileExist(csvFile, func(filePath string) (bool, error) {
		panic(horus.NewHerror("cloneRepositoriesFromCSV", "CSV file does not exist", nil,
			map[string]any{"csvFile": filePath}))
	}, true)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to check existence of CSV file: "+csvFile)
	}

	file, err := os.Open(csvFile)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to open CSV file: "+csvFile)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to parse CSV file: "+csvFile)
	}

	start := 0
	if len(records) > 0 && strings.EqualFold(records[0][0], "repoName") {
		start = 1
	}

	for i := start; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 2 {
			continue
		}
		repoName := strings.TrimSpace(rec[0])
		repoURL := strings.TrimSpace(rec[1])
		if repoName == "" || repoURL == "" {
			continue
		}
		finalTargetDir := filepath.Join(targetDir, repoName)
		if err := cloneRepository(repoURL, finalTargetDir); err != nil {
			return horus.Wrap(err, "cloneRepositoriesFromCSV",
				"failed to clone repository at row "+strconv.Itoa(i+1))
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
