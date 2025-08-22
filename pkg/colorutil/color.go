package colorutil

import "github.com/fatih/color"

var (
	Green = color.New(color.FgGreen).PrintfFunc()
	Red   = color.New(color.FgRed).PrintfFunc()
	Cyan  = color.New(color.FgCyan).PrintfFunc()
)
