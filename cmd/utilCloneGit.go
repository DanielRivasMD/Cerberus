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

// cloneRepositoriesFromCSV reads a CSV file whose rows contain Git repository details,
// where the first column is the repository name and the second column is the repository URL.
// It then clones each repository into a subdirectory under targetDir using the provided repository name.
// If targetDir is an empty string, it defaults to "." (the current directory).
// Before opening the CSV file, it verifies its existence using domovoi.FileExist with an anonymous NotFoundAction.
func cloneRepositoriesFromCSV(csvFile, targetDir string) error {
	// Set default targetDir if not provided.
	if strings.TrimSpace(targetDir) == "" {
		targetDir = "."
	}

	// Verify that the CSV file exists.
	// Anonymous function panics on not found scenario
	_, err := domovoi.FileExist(csvFile, func(filePath string) (bool, error) {
		panic(horus.NewHerror("cloneRepositoriesFromCSV", "CSV file does not exist", nil, map[string]any{"csvFile": filePath}))
	}, true)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to check existence of CSV file: "+csvFile)
	}

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

	// Process each row. Expecting two columns: [repoName, repoURL].
	for i, record := range records {
		// Check that there are at least two columns.
		if len(record) < 2 {
			continue
		}

		repoName := strings.TrimSpace(record[0])
		repoURL := strings.TrimSpace(record[1])

		// Skip if either field is empty.
		if repoName == "" || repoURL == "" {
			continue
		}

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
