////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
  "bufio"
  // "fmt"
  "os"
  "strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// parseReadme extracts the content under "### Description"
func parseReadme(filename string) (string, error) {
  file, err := os.Open(filename)
  if err != nil {
    return "", err
  }
  defer file.Close()

  var descriptionLines []string
  scanner := bufio.NewScanner(file)
  inDescription := false

  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())

    // Check if we've reached "### Description"
    if strings.HasPrefix(line, "### Description") {
      inDescription = true
      continue
    }

    // Stop reading when we hit another heading (###)
    if inDescription && strings.HasPrefix(line, "### ") {
      break
    }

    if inDescription {
      descriptionLines = append(descriptionLines, line)
    }
  }

  if err := scanner.Err(); err != nil {
    return "", err
  }

  return strings.Join(descriptionLines, "\n"), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
