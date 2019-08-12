package main

import (
	"fmt"
)

func tabularOutput(data [][]string) {

	if len(data) > 0 {
		columnCount := len(data[0])

		columnWidths := make([]int, columnCount)

		for row := 0; row < len(data); row++ {
			for col := 0; col < columnCount; col++ {
				if len(data[row][col]) > columnWidths[col] {
					if len(data[row][col]) > 50 {
						columnWidths[col] = 50
					} else {
						columnWidths[col] = len(data[row][col])
					}
				}
			}
		}

		outputFormatSpecs := make([]string, columnCount)
		for col := 0; col < columnCount; col++ {
			outputFormatSpecs[col] = fmt.Sprintf("%%-%ds", columnWidths[col])
		}

		// Do the actual printing
		for row := 0; row < len(data); row++ {
			for col := 0; col < columnCount; col++ {
				fmt.Printf(outputFormatSpecs[col], data[row][col])
				fmt.Printf("  ") // two-space gutter between columns
			}
			fmt.Printf("\n") // end of each row
		}
	}
}

type boolProperty struct {
	Name  string
	Value bool
}
