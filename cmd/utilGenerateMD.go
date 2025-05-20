////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// getHeaders extracts the field names of a struct,
// omitting any fields present in skipFields.
func getHeaders(v interface{}, skipFields map[string]bool) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var headers []string
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if skipFields[fieldName] {
			continue // skip the field if it's in the skipFields map
		}
		headers = append(headers, fieldName)
	}
	return headers
}

// getValues returns the string representations of the struct’s field values,
// formatted with fixed widths taken from the fieldSizes slice,
// and omits any fields that are in skipFields.
func getValues(v interface{}, fieldSizes []int, skipFields map[string]bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var values []string
	col := 0
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if skipFields[fieldName] {
			continue // skip the field if it's in the skipFields map
		}
		width := 0
		if col < len(fieldSizes) {
			width = fieldSizes[col]
		}
		formatted := fmt.Sprintf("%-*v", width, val.Field(i).Interface())
		values = append(values, formatted)
		col++
	}
	return values
}

// generateMarkdownHeader creates a Markdown header row including a separator.
// It uses the provided fieldSizes slice to pad each header and accepts
// a skipFields map so that specified fields are not printed.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row
	builder.WriteString("| ")
	for i, header := range headers {
		width := 0
		if i < len(fieldSizes) {
			width = fieldSizes[i]
		}
		if width > 0 {
			builder.WriteString(fmt.Sprintf("%-*s | ", width, header))
		} else {
			builder.WriteString(header + " | ")
		}
	}
	builder.WriteString("\n")

	// Separator row
	builder.WriteString("|")
	for i, header := range headers {
		width := fieldSizes[i]
		if width <= 0 {
			width = len(header)
		}
		builder.WriteString(strings.Repeat("-", width+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMarkdownRow creates a Markdown table row for a single struct instance.
// It accepts a skipFields map so that fields (e.g., "Remote", "Files", "Frequency")
// can be omitted from the output.
	values := getValues(v, fieldSizes, skipFields)
	var builder strings.Builder

	builder.WriteString("| ")
	for i, value := range values {
		if i < len(fieldSizes) && fieldSizes[i] > 0 {
			builder.WriteString(fmt.Sprintf("%-*s | ", fieldSizes[i], value))
		} else {
			builder.WriteString(value + " | ")
		}
	}
	builder.WriteString("\n")
	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// getColoredLanguage pads the input language string to the provided width,
// then applies color based on its base language (first token).
func getColoredLanguage(language string, width int) string {
	// First, pad the entire language string to the required width.
	padded := fmt.Sprintf("%-"+fmt.Sprintf("%d", width)+"s", language)
	// Split the language string by spaces to determine the base language.
	parts := strings.Split(language, " ")
	baseLanguage := strings.ToLower(parts[0])
	switch baseLanguage {
	case "go":
		return chalk.Blue.Color(padded)
	case "julia":
		return chalk.Magenta.Color(padded)
	case "python":
		return chalk.Green.Color(padded)
	case "r":
		return chalk.Cyan.Color(padded)
	case "rust":
		return chalk.Yellow.Color(padded)
	case "shell":
		return chalk.Red.Color(padded)
	default:
		return padded // Return uncolored if no match.
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// calculateRepoAgeInMonths calculates repository age in months given an "Xy Ym" format.
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // Parse "Xy Ym" format.
	return (years * 12) + months
}

////////////////////////////////////////////////////////////////////////////////////////////////////
