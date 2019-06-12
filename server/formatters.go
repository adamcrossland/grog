package main

import (
	"fmt"
	"io"
	"strconv"
	"time"

	model "github.com/adamcrossland/grog/models"
	"github.com/adamcrossland/grog/mtemplate"
)

// ShortDateFormatter formats a Time value into a brief representation.
func ShortDateFormatter(w io.Writer, format string, data *mtemplate.TemplateData, value ...interface{}) {
	var foundTime bool
	var timeToRender time.Time

	if len(value) > 0 {
		timeVal, timeValOK := value[0].(model.NullTime)
		if timeValOK {
			timeToRender = timeVal.Val()
			foundTime = true
		} else {
			// Try to convert from an int64 representation of time
			intVal, intValOK := value[0].(int64)
			if intValOK {
				timeToRender = time.Unix(intVal, 0)
				foundTime = true
			} else {
				// Try to convert from a string value of an int representation of time
				strVal, strValOK := value[0].(string)
				if strValOK {
					parsedIntVal, parsedIntValErr := strconv.ParseInt(strVal, 10, 64)
					if parsedIntValErr == nil {
						timeToRender = time.Unix(parsedIntVal, 0)
						foundTime = true
					}
				}
			}
		}
	}

	if foundTime {
		fmt.Fprintf(w, "%s %d %d", timeToRender.Month(), timeToRender.Day(), timeToRender.Year())
		return
	}

	// Could not convert this value to a date, so just output it.
	fmt.Fprint(w, value...)
}
