////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func commitFrequency(year int) (map[string]int, error) {
	// initialize map with all months set 0
	commitFrequency := map[string]int{
		fmt.Sprintf("%d-01", year): 0,
		fmt.Sprintf("%d-02", year): 0,
		fmt.Sprintf("%d-03", year): 0,
		fmt.Sprintf("%d-04", year): 0,
		fmt.Sprintf("%d-05", year): 0,
		fmt.Sprintf("%d-06", year): 0,
		fmt.Sprintf("%d-07", year): 0,
		fmt.Sprintf("%d-08", year): 0,
		fmt.Sprintf("%d-09", year): 0,
		fmt.Sprintf("%d-10", year): 0,
		fmt.Sprintf("%d-11", year): 0,
		fmt.Sprintf("%d-12", year): 0,
	}

	// get commit dates within specified year
	out, _, ε := domovoi.CaptureExecCmd("git", "log", "--since", fmt.Sprintf("%d-01-01", year), "--until", fmt.Sprintf("%d-12-31", year), "--format=%ci")
	horus.CheckErr(ε)

	// process output & group by month
	commitDates := strings.Split(string(out), "\n")
	layout := "2006-01-02 15:04:05 -0700" // git date format

	for _, dateStr := range commitDates {
		if strings.TrimSpace(dateStr) == "" {
			continue // skip empty lines
		}

		commitTime, err := time.Parse(layout, dateStr)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		// use "YYYY-MM" format for grouping by month
		month := commitTime.Format("2006-01")
		commitFrequency[month]++
	}

	return commitFrequency, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
