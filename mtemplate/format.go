// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Template library: default formatters

package mtemplate

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// FormatterFunc is the signature which a function intended to be an mtemplate
// formatter must follow
type FormatterFunc func(io.Writer, string, *TemplateData, ...interface{})

// StringFormatter formats into the default string representation.
// It is stored under the name "str" and is the default formatter.
// You can override the default formatter by storing your default
// under the name "" in your custom formatter map.
func StringFormatter(w io.Writer, format string, data *TemplateData, value ...interface{}) {
	if len(value) == 1 {
		if b, ok := value[0].([]byte); ok {
			w.Write(b)
			return
		}
	}
	fmt.Fprint(w, value...)
}

// IntFormatter formats into an integer representation.
func IntFormatter(w io.Writer, format string, data *TemplateData, value ...interface{}) {
	if len(value) == 1 {
		strVal, strValOK := value[0].(string)
		if !strValOK || len(strVal) > 0 {
			// Only emit if there is a non-empty value to convert or the value
			// is something other than a string -- certainly an int in this case.
			fmt.Fprintf(w, "%d", value[0])
		}
		return
	}
	fmt.Fprint(w, value...)
}

var (
	escQuot = []byte("&#34;") // shorter than "&quot;"
	escApos = []byte("&#39;") // shorter than "&apos;"
	escAmp  = []byte("&amp;")
	escLt   = []byte("&lt;")
	escGt   = []byte("&gt;")
)

// HTMLEscape writes to w the properly escaped HTML equivalent
// of the plain text data s.
func HTMLEscape(w io.Writer, s []byte) {
	var esc []byte
	last := 0
	for i, c := range s {
		switch c {
		case '"':
			esc = escQuot
		case '\'':
			esc = escApos
		case '&':
			esc = escAmp
		case '<':
			esc = escLt
		case '>':
			esc = escGt
		default:
			continue
		}
		w.Write(s[last:i])
		w.Write(esc)
		last = i + 1
	}
	w.Write(s[last:])
}

// HTMLFormatter formats arbitrary values for HTML
func HTMLFormatter(w io.Writer, format string, data *TemplateData, value ...interface{}) {
	ok := false
	var b []byte
	if len(value) == 1 {
		b, ok = value[0].([]byte)
	}
	if !ok {
		var buf bytes.Buffer
		fmt.Fprint(&buf, value...)
		b = buf.Bytes()
	}
	HTMLEscape(w, b)
}

// URLFormatter formats arbitrary values for inclusion in URL
// paramters
func URLFormatter(w io.Writer, format string, data *TemplateData, value ...interface{}) {
	asString := ""

	if len(value) >= 1 {
		if b, ok := value[0].([]byte); ok {
			var inBuffer bytes.Buffer
			inBuffer.Write(b)
			asString = inBuffer.String()
		} else {
			asString = fmt.Sprint(value...)
		}
	}
	asString = strings.Trim(asString, "[]")
	asString = strings.TrimSpace(asString)
	safeString := url.QueryEscape(asString)
	safeAsBuffer := bytes.NewBufferString(safeString)
	w.Write(safeAsBuffer.Bytes())
}

func pageKey(key string) string {
	return key + "-page"
}

func pageNextPageKey(key string) string {
	return key + "-next-page"
}

func pagePrevPageKey(key string) string {
	return key + "-prev-page"
}

func pageTotalPagesKey(key string) string {
	return key + "-total-pages"
}

// PaginationFormatter takes an input array and resizes it to the number and set
// of elements to be displayed. Adds several cookies and data elements to all
// pagination to persist across page views and to allow pagination controls
// to be rendered.
func PaginationFormatter(w io.Writer, format string, data *TemplateData, value ...interface{}) {
	params := strings.Split(format, " ")
	if len(params) < 3 {
		panic("pagination formatter must have at least two parameters: paginationkey and pagesize. a third parameter, currentpage, is optional")
	}

	key := params[1]
	pageSize, pageSizeErr := strconv.ParseInt(params[2], 10, 32)
	if pageSizeErr != nil {
		panic(fmt.Sprintf("paginationFormatter: second parameter (%s) must be convertible to an integer", params[2]))
	}

	var showPage int64

	pageRequested := data.request.FormValue(pageKey(key))
	if pageRequested != "" {
		showPage, _ = strconv.ParseInt(pageRequested, 10, 32)
	}

	// The real default value is 1
	if showPage == 0 {
		showPage = 1
	}

	realData := value[0].([]map[string]string)

	// Calculate how many potential pages there are.
	totalPages := int64(len(realData)) / pageSize
	if int64(len(realData))%pageSize != 0 {
		totalPages++
	}

	dataOffset := (showPage - 1) * pageSize
	resultCount := int64(len(realData[dataOffset:]))
	if resultCount < pageSize {
		pageSize = resultCount
	}

	paginatedData := make([]map[string]string, pageSize)

	copy(paginatedData, realData[dataOffset:])
	data.data[pageKey(key)] = paginatedData

	data.data[pageTotalPagesKey((key))] = totalPages

	if showPage > 1 {
		data.data[pagePrevPageKey(key)] = showPage - 1
	}

	if showPage < totalPages {
		data.data[pageNextPageKey(key)] = showPage + 1
	}
}
