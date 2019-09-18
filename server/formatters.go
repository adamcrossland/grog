package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
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

// TruncFormatter shortens the input data to a maximum length and appends an optional string to
// indicate that the data has been truncated
func TruncFormatter(w io.Writer, format string, data *mtemplate.TemplateData, value ...interface{}) {
	params := getStringFromQuotes(format)

	if len(params) >= 2 {
		truncToLength, lengthErr := strconv.ParseInt(params[1], 10, 64)
		if lengthErr == nil {
			var asString string

			if b, ok := value[0].([]byte); ok {
				var inBuffer bytes.Buffer
				inBuffer.Write(b)
				asString = inBuffer.String()
			} else if asString, ok = value[0].(string); !ok {
				// Cannot convert the input to a truncatable string; write it out and exit
				fmt.Fprint(w, value...)
				return
			}

			if int64(len(asString)) > truncToLength {
				fmt.Fprint(w, asString[0:truncToLength])

				if len(params) == 3 {
					ellipsisText := params[2]
					// If the parameter is quoted, we should remove the quotes.
					if ellipsisText[0] == '"' {
						ellipsisText = ellipsisText[1:]
					}
					if ellipsisText[len(ellipsisText)-1] == '"' {
						ellipsisText = ellipsisText[:len(ellipsisText)-1]
					}

					fmt.Fprint(w, ellipsisText)
				}
			} else {
				// The input data is shorter than the desired length, so just write it out.
				fmt.Fprint(w, value...)
			}

		} else {
			panic(fmt.Sprintf("trunc formatter: param 1 (%s) must be convertible to int; failed: %v", params[0], lengthErr))
		}
	} else {
		panic("trunc formatter: requires at least 1 parameter")
	}
}

// getStringFromQuotes finds a "string which spans multiple spaces" in a split message.
// Then takes that and replaces the Quote string with a single string value of the quote contents
// credit to https://scene-si.org/2017/09/02/parsing-strings-with-go/
func getStringFromQuotes(toParse string) []string {
	inQuote := false
	f := func(c rune) bool {
		switch {
		case c == '"':
			inQuote = !inQuote
			return false
		case inQuote:
			return false
		default:
			return c == ' '
		}
	}
	return strings.FieldsFunc(toParse, f)
}
