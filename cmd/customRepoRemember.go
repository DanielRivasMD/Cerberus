////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	// "os"

	// "github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoRemember represents the repository features.
type RepoRemember struct {
	Repo   string // Repo name
	Remote string // Remote URL of the repository
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateRepoRemember gathers information about a repository,
// wrapping any errors using the Horus library for additional context.
func populateRepoRemember() (RepoRemember, error) {
	// initialize RepoRemember
	remember := RepoRemember{}

	remoteURL, err := getRemote()
	if err != nil {
		return remember, horus.Wrap(err, "populateRepoRemember", "failed to get remote repository address")
	}

	remember.Remote = remoteURL

	return remember, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateRememberMD generates the Markdown table for the describe command.
func generateRememberMD(repoNames []string) string {
	// Define column widths for: Repo, Remote
	fieldSizes := []int{Defaults.repoLen, Defaults.remoteLen}
	skip := map[string]bool{}

	var sample RepoRemember // used solely for header generation

	// populateFunc for describe command.
	populateFunc := func(repoName string) (*RepoRemember, error) {
		d, err := populateRepoRemember()
		if err != nil {
			return nil, err
		}
		d.Repo = repoName
		return &d, nil
	}

	// Create an aligners map that forces all fields to be left aligned.
	aligners := map[string]Alignment{
		"Repo":   AlignLeft,
		"Remote": AlignLeft,
	}

	// Extra parameter is not used for describe (pass 0).
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, 0, aligners, output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
