/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ttacon/chalk"
)

// ----------------------------------------------------------------------
// Main entry point
// ----------------------------------------------------------------------

func executeLs(args []string) {
	if len(args) == 0 {
		args = []string{"."}
	}

	if lsOpts.directory && lsOpts.noDirectory {
		fmt.Fprintf(os.Stderr, "Error: --directory and --no-directory cannot be used together\n")
		os.Exit(1)
	}

	sortSpec, reverse := resolveSortSpec()

	for i, path := range args {
		if len(args) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", path)
		}
		listPath(path, sortSpec, reverse)
	}
}

// resolveSortSpec returns a string like "name", "size", "mtime", etc., and reverse flag.
// It implements the logic of the original k's case statement.
func resolveSortSpec() (string, bool) {
	// Priority: explicit --sort flag overrides short options
	if lsOpts.sortWord != "" {
		switch lsOpts.sortWord {
		case "none":
			return "none", lsOpts.reverse
		case "size":
			return "size", lsOpts.reverse
		case "time":
			return "mtime", lsOpts.reverse
		case "ctime", "status":
			return "ctime", lsOpts.reverse
		case "atime", "access", "use":
			return "atime", lsOpts.reverse
		default:
			// fallback to name
			return "name", lsOpts.reverse
		}
	}

	// Short option logic
	switch {
	case lsOpts.unsorted:
		return "none", lsOpts.reverse
	case lsOpts.sortBySize:
		return "size", lsOpts.reverse
	case lsOpts.sortByTime:
		return "mtime", lsOpts.reverse
	case lsOpts.sortByCtime:
		return "ctime", lsOpts.reverse
	case lsOpts.sortByAtime:
		return "atime", lsOpts.reverse
	default:
		return "name", lsOpts.reverse
	}
}

// ----------------------------------------------------------------------
// Path listing
// ----------------------------------------------------------------------

type fileEntry struct {
	name     string
	fullPath string
	info     os.FileInfo
	stat     *syscall.Stat_t // for detailed metadata
	err      error
}

func listPath(path string, sortSpec string, reverse bool) {
	// Check if path is a file
	fi, err := os.Lstat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "k: cannot access %s: %v\n", path, err)
		return
	}
	if !fi.IsDir() {
		// Single file: just list it with details
		entries := []fileEntry{{
			name:     fi.Name(),
			fullPath: path,
			info:     fi,
			stat:     statFromFileInfo(fi),
		}}
		printEntries(entries, path)
		return
	}

	// Directory: read entries
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "k: cannot read directory %s: %v\n", path, err)
		return
	}

	// Filter and collect entries
	var entries []fileEntry
	for _, de := range dirEntries {
		name := de.Name()
		fullPath := filepath.Join(path, name)

		// Filter dot files
		if strings.HasPrefix(name, ".") {
			if !lsOpts.all && !lsOpts.almostAll {
				continue
			}
			if lsOpts.almostAll && (name == "." || name == "..") {
				continue
			}
		}

		info, err := de.Info()
		if err != nil {
			// Still include but with error
			entries = append(entries, fileEntry{name: name, fullPath: fullPath, err: err})
			continue
		}

		// Apply directory filters
		isDir := info.IsDir()
		if lsOpts.directory && !isDir {
			continue
		}
		if lsOpts.noDirectory && isDir {
			continue
		}

		// Collect metadata
		stat := statFromFileInfo(info)
		entries = append(entries, fileEntry{
			name:     name,
			fullPath: fullPath,
			info:     info,
			stat:     stat,
		})
	}

	// Sort entries
	sortEntries(entries, sortSpec, reverse)

	// Add . and .. if -a (and not -A) and directory
	if lsOpts.all && !lsOpts.almostAll && !lsOpts.directory && !lsOpts.noDirectory {
		// We need to get stats for . and .. as well
		dotEntries := []string{".", ".."}
		for _, dot := range dotEntries {
			fullPath := filepath.Join(path, dot)
			info, err := os.Lstat(fullPath)
			if err != nil {
				continue
			}
			stat := statFromFileInfo(info)
			// Insert at beginning, unsorted
			entries = append([]fileEntry{{
				name:     dot,
				fullPath: fullPath,
				info:     info,
				stat:     stat,
			}}, entries...)
		}
	}

	// Print
	printEntries(entries, path)
}

// statFromFileInfo attempts to get syscall.Stat_t from FileInfo.
func statFromFileInfo(info os.FileInfo) *syscall.Stat_t {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return sys
	}
	return nil
}

// ----------------------------------------------------------------------
// Sorting
// ----------------------------------------------------------------------

func sortEntries(entries []fileEntry, sortSpec string, reverse bool) {
	if sortSpec == "none" {
		return
	}

	less := func(i, j int) bool {
		var cmp bool
		switch sortSpec {
		case "name":
			cmp = entries[i].name < entries[j].name
		case "size":
			cmp = entries[i].info.Size() < entries[j].info.Size()
		case "mtime":
			cmp = entries[i].info.ModTime().Before(entries[j].info.ModTime())
		case "ctime":
			ci := ctime(entries[i].stat)
			cj := ctime(entries[j].stat)
			cmp = ci.Before(cj)
		case "atime":
			ai := atime(entries[i].stat)
			aj := atime(entries[j].stat)
			cmp = ai.Before(aj)
		default:
			cmp = entries[i].name < entries[j].name
		}
		if reverse {
			return !cmp
		}
		return cmp
	}
	sort.Slice(entries, less)
}

func atime(stat *syscall.Stat_t) time.Time {
	if stat == nil {
		return time.Time{}
	}
	switch runtime.GOOS {
	case "darwin":
		// macOS uses Atimespec and Ctimespec
		return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
	// case "linux", "freebsd", "netbsd", "openbsd", "dragonfly":
	// Linux and most BSDs use Atim and Ctim
	// return time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	default:
		// Unsupported platform – return zero time (sorting by atime will be stable but meaningless)
		return time.Time{}
	}
}

func ctime(stat *syscall.Stat_t) time.Time {
	if stat == nil {
		return time.Time{}
	}
	switch runtime.GOOS {
	case "darwin":
		return time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
	// case "linux", "freebsd", "netbsd", "openbsd", "dragonfly":
	// return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	default:
		return time.Time{}
	}
}

// ----------------------------------------------------------------------
// Output formatting
// ----------------------------------------------------------------------

// Column widths
type colWidths struct {
	mode  int
	nlink int
	owner int
	group int
	size  int
}

func printEntries(entries []fileEntry, basePath string) {
	if len(entries) == 0 {
		return
	}

	// First pass: collect metadata and compute max widths
	widths := colWidths{}
	fileInfos := make([]*fileInfoDisplay, len(entries))

	// For total blocks calculation (like ls)
	var totalBlocks int64

	for i, e := range entries {
		if e.err != nil {
			// Skip problematic entries for now
			fileInfos[i] = &fileInfoDisplay{name: e.name, err: e.err}
			continue
		}
		info := e.info
		stat := e.stat

		// Mode string (permissions)
		modeStr := info.Mode().String()
		if len(modeStr) > widths.mode {
			widths.mode = len(modeStr)
		}

		// Number of links
		var nlink uint64
		if stat != nil {
			nlink = uint64(stat.Nlink)
		}
		nlinkStr := strconv.FormatUint(nlink, 10)
		if len(nlinkStr) > widths.nlink {
			widths.nlink = len(nlinkStr)
		}

		// Owner and group names
		owner := username(stat)
		group := groupname(stat)
		if len(owner) > widths.owner {
			widths.owner = len(owner)
		}
		if len(group) > widths.group {
			widths.group = len(group)
		}

		// Size (human or raw)
		size := info.Size()
		totalBlocks += blocks(info, stat)
		sizeStr := formatSize(size, lsOpts.human, lsOpts.si)
		if len(sizeStr) > widths.size {
			widths.size = len(sizeStr)
		}

		fileInfos[i] = &fileInfoDisplay{
			name:     e.name,
			fullPath: e.fullPath,
			mode:     modeStr,
			nlink:    nlinkStr,
			owner:    owner,
			group:    group,
			size:     size,
			sizeStr:  sizeStr,
			mtime:    info.ModTime(),
			stat:     stat,
			info:     info,
		}
	}

	// Print total blocks line (like "total 8")
	fmt.Printf("total %d\n", totalBlocks)

	// Second pass: print each entry with git status
	for _, fi := range fileInfos {
		if fi == nil || fi.err != nil {
			if fi != nil {
				fmt.Fprintf(os.Stderr, "k: error accessing %s: %v\n", fi.name, fi.err)
			}
			continue
		}

		// Get git status if not disabled
		gitMarker := ""
		if !lsOpts.noVCS {
			gitMarker = gitStatus(fi.fullPath, fi.info.IsDir())
		}

		// Format each field with padding
		modeField := fmt.Sprintf("%-*s", widths.mode, fi.mode)
		nlinkField := fmt.Sprintf("%*s", widths.nlink, fi.nlink)
		ownerField := fmt.Sprintf("%-*s", widths.owner, fi.owner)
		groupField := fmt.Sprintf("%-*s", widths.group, fi.group)
		sizeField := fmt.Sprintf("%*s", widths.size, fi.sizeStr)

		// Date and time
		dateField := formatDate(fi.mtime)

		// Colored name
		nameField := colorName(fi.name, fi.info)

		// Symlink target if any
		symlinkTarget := ""
		if fi.info.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Readlink(fi.fullPath); err == nil {
				symlinkTarget = " -> " + target
			}
		}

		// Assemble line
		line := fmt.Sprintf("%s %s %s %s %s %s %s%s%s",
			modeField, nlinkField, ownerField, groupField,
			sizeField, dateField, gitMarker, nameField, symlinkTarget)
		fmt.Println(line)
	}
}

type fileInfoDisplay struct {
	name     string
	fullPath string
	err      error
	mode     string
	nlink    string
	owner    string
	group    string
	size     int64
	sizeStr  string
	mtime    time.Time
	stat     *syscall.Stat_t
	info     os.FileInfo
}

// blocks returns the number of 512-byte blocks used by the file (like ls).
func blocks(info os.FileInfo, stat *syscall.Stat_t) int64 {
	if stat != nil {
		// stat.Blocks is in 512-byte units
		return int64(stat.Blocks)
	}
	// Fallback: size / 512, rounded up
	return (info.Size() + 511) / 512
}

// username returns the user name from uid, or the uid as string.
func username(stat *syscall.Stat_t) string {
	if stat == nil {
		return "?"
	}
	u, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
	if err != nil {
		return strconv.Itoa(int(stat.Uid))
	}
	return u.Username
}

// groupname returns the group name from gid, or the gid as string.
func groupname(stat *syscall.Stat_t) string {
	if stat == nil {
		return "?"
	}
	g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
	if err != nil {
		return strconv.Itoa(int(stat.Gid))
	}
	return g.Name
}

// formatSize converts size to human-readable if requested.
func formatSize(size int64, human, si bool) string {
	if !human {
		return strconv.FormatInt(size, 10)
	}
	units := []string{"B", "K", "M", "G", "T", "P", "E"}
	base := int64(1024)
	if si {
		base = 1000
	}
	if size < base {
		return fmt.Sprintf("%d%s", size, units[0])
	}
	div, exp := base, 0
	for n := size / base; n >= base; n /= base {
		div *= base
		exp++
	}
	return fmt.Sprintf("%.1f%s", float64(size)/float64(div), units[exp+1])
}

// formatDate returns a string like "Jan  2 15:04" or "Jan  2  2006" depending on age.
func formatDate(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() {
		// Within this year: show time
		return t.Format("Jan _2 15:04")
	}
	// Older: show year
	return t.Format("Jan _2  2006")
}

// colorName applies ANSI color based on file type and permissions, mimicking LSCOLORS.
func colorName(name string, info os.FileInfo) string {
	mode := info.Mode()
	// Default to no color
	colored := name

	// Determine color code (simplified version of LSCOLORS mapping)
	// For now, use hardcoded colors similar to k's defaults:
	// di=34 (blue), ln=35 (magenta), so=32 (green), pi=33 (yellow), ex=31 (red)
	switch {
	case mode.IsDir():
		colored = chalk.Blue.Color(name)
	case mode&os.ModeSymlink != 0:
		colored = chalk.Magenta.Color(name)
	case mode&os.ModeSocket != 0:
		colored = chalk.Green.Color(name)
	case mode&os.ModeNamedPipe != 0:
		colored = chalk.Yellow.Color(name)
	case mode&0100 != 0: // executable for owner
		colored = chalk.Red.Color(name)
	default:
		// regular file: no color
	}
	return colored
}

// gitStatus returns a marker indicating the file's Git status.
// For simplicity, we call `git status --porcelain` on the file.
func gitStatus(path string, isDir bool) string {
	// Determine the git top-level directory
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = filepath.Dir(path)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	topLevel := strings.TrimSpace(string(out))
	if topLevel == "" {
		return ""
	}

	// Get relative path
	rel, err := filepath.Rel(topLevel, path)
	if err != nil {
		return ""
	}
	if isDir {
		rel += "/"
	}

	// Run git status --porcelain
	cmd = exec.Command("git", "status", "--porcelain", "--", rel)
	cmd.Dir = topLevel
	out, err = cmd.Output()
	if err != nil {
		return ""
	}
	status := strings.TrimSpace(string(out))
	if status == "" {
		// No changes
		return " "
	}
	// First two characters indicate status
	if len(status) >= 2 {
		code := status[:2]
		// Map to marker symbols (as in k)
		switch code {
		case " M": // modified, not staged
			return chalk.Red.Color("+")
		case "M ": // modified, staged
			return chalk.Green.Color("+")
		case "??": // untracked
			return chalk.Color(chalk.Yellow).Color("+") // approximate orange
		case "!!": // ignored
			return chalk.Dim.TextStyle("|")
		default:
			return " "
		}
	}
	return " "
}
