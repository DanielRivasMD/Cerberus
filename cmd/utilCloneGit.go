////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// cloneGit clones a Git repository from the specified URL into the target directory.
// It wraps any errors using horus.Wrap to provide context.
func cloneGit(repoURL, targetDir string) error {
	out, _, err := domovoi.CaptureExecCmd("git", "clone", repoURL, targetDir)
	if err != nil {
		return horus.Wrap(err, "cloneRepository", "failed to clone repository: "+repoURL)
	}

	// Optionally, process output if needed (for example, logging the output).
	_ = strings.TrimSpace(string(out))

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
