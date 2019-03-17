package main

import (
	"fmt"
	"io"

	"github.com/adamcrossland/grog/models"
)

// ShortDateFormatter formats a Time value into a brief representation.
func ShortDateFormatter(w io.Writer, format string, value ...interface{}) {
	if len(value) == 1 {
		timeVal, timeValOK := value[0].(model.NullTime)
		if timeValOK {
			fmt.Fprintf(w, "%s %d %d", timeVal.Val().Month(), timeVal.Val().Day(), timeVal.Val().Year())
			return
		}
	}

	fmt.Fprint(w, value...)
}
