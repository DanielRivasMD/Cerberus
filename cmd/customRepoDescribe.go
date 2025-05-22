////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoDescribe represents the repository features.
type RepoDescribe struct {
	Repo     string
	Remote   string
	Overview string
	License  string
}

// TODO: replace manual header generator
// generateHeader dynamically creates a Markdown table header
// based on the field names of the provided struct.
func generateHeader(v interface{}) string {
	// Get the underlying type (in case a pointer is passed).
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var builder strings.Builder

	// Build the header row.
	builder.WriteString("|")
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		builder.WriteString(" " + fieldName + " |")
	}
	builder.WriteString("\n")

	// Build the separator row.
	builder.WriteString("|")
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		// Create dashes with a count equal to the length of the field name plus some padding.
		dashCount := len(fieldName) + 2
		builder.WriteString(strings.Repeat("-", dashCount) + "|")
	}
	builder.WriteString("\n")

	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateHeaderDescribe creates the Markdown table header for the RepoDescribe fields.
func generateHeaderDescribe() string {
	var builder strings.Builder
	// Modify the header row formatting (column widths) as needed.
	builder.WriteString("| License          | Overview                       | Remote                   |\n")
	builder.WriteString("|------------------|--------------------------------|--------------------------|\n")
	return builder.String()
}

// generateBodyDescribe creates a Markdown table row for a single repository's description.
func generateBodyDescribe(repo RepoDescribe) string {
	// Use sprintf formatting to enforce fixed-width columns.
	// Adjust the widths per your display requirements.
	return fmt.Sprintf("| %-16s | %-30s | %-24s |\n", repo.License, repo.Overview, repo.Remote)
}

// generateMDDescribe creates the complete Markdown table for multiple repositories.
func generateMDDescribe(repos []RepoDescribe) string {
	var builder strings.Builder
	// Add the Markdown header once.
	builder.WriteString(generateHeaderDescribe())

	// Iterate over the repositories and add a row for each.
	for _, repo := range repos {
		builder.WriteString(generateBodyDescribe(repo))
	}

	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: use horus to catch errors
// TODO: find why multiple repos error out
func populateRepoDescribe() (RepoDescribe, error) {
	// initialize RepoDescribe
	describe := RepoDescribe{}

	// list files
	files, err := listFiles(repository)
	if err != nil {
		return describe, err
	}

	// declare switches
	readmeFound := false
	licenseFound := false

	// iterate on files
	for _, file := range files {
		if file == "README.md" {
			readmeFound = true
			describe.Overview, err = parseReadme(file)
			if err != nil {
				return describe, err
			}
			fmt.Println("Extracted Description:\n", describe.Overview)
		}

		if file == "LICENSE" {
			licenseFound = true
			describe.License, err = detectLicense(file)
			if err != nil {
				return describe, err
			}
			fmt.Println("License is: ", describe.License)
		}
	}

	if !readmeFound {
		fmt.Println("README.md not found in the directory.")
	}
	if !licenseFound {
		fmt.Println("LICENSE not found in the directory.")
	}

	// define remote
	remoteURL, err := getRemote()
	if err != nil {
		return describe, err
	}
	describe.Remote = remoteURL

	fmt.Println(describe.Remote)

	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
