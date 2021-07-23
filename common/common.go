package common

import (
	"regexp"

	"github.com/360EntSecGroup-Skylar/excelize"
)

var ReDigit = regexp.MustCompile(`\D`)

var Xlsx = excelize.NewFile()
var abs = []rune("ABCDEFGHIJKLMNOPQRSTUVWXUZ")
