package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func formIntValueOrDef(r *http.Request, formKey string) int64 {
	var foundValue int64

	valStr, valOK := r.Form[formKey]

	if valOK && len(valStr[0]) > 0 {
		valInt, valIntOK := strconv.Atoi(valStr[0])
		if valIntOK == nil {
			foundValue = int64(valInt)
		}
	}

	return foundValue
}

func formStringValueOrDef(r *http.Request, formKey string) string {
	var foundValue string

	valStr, valOK := r.Form[formKey]

	if valOK && len(valStr[0]) > 0 {
		foundValue = valStr[0]
	}

	return foundValue
}

func formIntValRequired(w http.ResponseWriter, r *http.Request, formKey string) int64 {
	result := formIntValueOrDef(r, formKey)

	if result == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s must be provided and cannot be 0", formKey)
		return 0
	}

	return result
}

func formStringValRequired(w http.ResponseWriter, r *http.Request, formKey string) string {
	result := formStringValueOrDef(r, formKey)

	if len(result) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s must be provided and cannot be empty", formKey)
		return result
	}

	return result
}
