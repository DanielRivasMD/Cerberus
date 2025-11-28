////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/DanielRivasMD/domovoi"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var exampleRoot = domovoi.FormatExample(
	"cerberus",
	[]string{"help"},
)

var exampleRemember = domovoi.FormatExample(
	"cerberus",
	[]string{"remember"},
	[]string{"remember", "--output", "repos.csv"},
)

var exampleReadme = domovoi.FormatExample(
	"cerberus",
	[]string{"readme"},
)

var exampleDescribe = domovoi.FormatExample(
	"cerberus",
	[]string{"describe"},
)

var exampleClone = domovoi.FormatExample(
	"cerberus",
	[]string{"clone"},
	[]string{"clone", "--csv", "repos.csv"},
)

var exampleRoadmap = domovoi.FormatExample(
	"cerberus",
	[]string{"roadmap"},
)

var exampleStats = domovoi.FormatExample(
	"cerberus",
	[]string{"stats"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////
