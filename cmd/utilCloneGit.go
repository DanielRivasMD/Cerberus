////////////////////////////////////////////////////////////////////////////////////////////////////

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

// cloneRepository clones a Git repository from the specified URL into the target directory.
// It wraps any errors using horus.Wrap to provide detailed context.
func cloneRepository(repoURL, targetDir string) error {
	out, _, err := domovoi.CaptureExecCmd("git", "clone", repoURL, targetDir)
	if err != nil {
		return horus.Wrap(err, "cloneRepository", "failed to clone repository: "+repoURL)
	}

	// Optionally process the output if needed.
	_ = strings.TrimSpace(string(out))
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// repoNameFromURL extracts the repository name from a given Git URL.
// For example, given "https://github.com/user/project.git", it returns "project".
func repoNameFromURL(repoURL string) string {
	base := filepath.Base(repoURL)
	base = strings.TrimSuffix(base, ".git")
	return base
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// cloneRepositoriesFromCSV reads a CSV file whose rows contain Git repository URLs,
// and clones each repository into a subdirectory under targetDir.
// Each repository is cloned into targetDir/<repoName>.
// Any errors encountered during file reading or cloning are wrapped using horus.Wrap.
func cloneRepositoriesFromCSV(csvFile, targetDir string) error {
	// Open the CSV file.
	file, err := os.Open(csvFile)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to open CSV file: "+csvFile)
	}
	defer file.Close()

	// Read all records from CSV.
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to parse CSV file: "+csvFile)
	}

	// Process each row.
	for i, record := range records {
		// Skip empty rows.
		if len(record) == 0 {
			continue
		}

		repoURL := strings.TrimSpace(record[0])
		if repoURL == "" {
			continue
		}

		// Determine the repository subdirectory name using repoNameFromURL.
		repoName := repoNameFromURL(repoURL)
		finalTargetDir := filepath.Join(targetDir, repoName)

		// Clone the repository.
		err := cloneRepository(repoURL, finalTargetDir)
		if err != nil {
			return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to clone repository at row "+strconv.Itoa(i+1))
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
