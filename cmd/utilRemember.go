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
	"strings"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateRememberCSV writes repo names and URLs to CSV.
func generateRememberCSV(repoNames []string) error {
	var outFile *os.File
	var err error

	if strings.TrimSpace(rootFlags.output) != "" {
		outFile, err = os.Create(rootFlags.output)
		if err != nil {
			return horus.Wrap(err, "generateRememberCSV", "failed to create output file: "+rootFlags.output)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	w := csv.NewWriter(outFile)
	defer w.Flush()

	if err := w.Write([]string{"repoName", "repoURL"}); err != nil {
		return err
	}
	for _, repo := range repoNames {
		remoteURL, err := resolveRemoteURL(repo)
		if err != nil {
			remoteURL = ""
		}
		if err := w.Write([]string{repo, remoteURL}); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
