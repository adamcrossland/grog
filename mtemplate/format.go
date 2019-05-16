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
	"strings"
)

// FormatterFunc is the signature which a function intended to be an mtemplate
// formatter must follow
type FormatterFunc func(io.Writer, string, ...interface{})

// StringFormatter formats into the default string representation.
// It is stored under the name "str" and is the default formatter.
// You can override the default formatter by storing your default
// under the name "" in your custom formatter map.
func StringFormatter(w io.Writer, format string, value ...interface{}) {
	if len(value) == 1 {
		if b, ok := value[0].([]byte); ok {
			w.Write(b)
			return
		}
	}
	fmt.Fprint(w, value...)
}

// IntFormatter formats into an integer representation.
func IntFormatter(w io.Writer, format string, value ...interface{}) {
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
func HTMLFormatter(w io.Writer, format string, value ...interface{}) {
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
func URLFormatter(w io.Writer, format string, value ...interface{}) {
	asString := ""

	if len(value) >= 1 {
		if b, ok := value[0].([]byte); ok {
			var inBuffer bytes.Buffer
			inBuffer.Write(b)
			asString = inBuffer.String()
		} else {
			asString = fmt.Sprint(value)
		}
	}
	asString = strings.Trim(asString, "[]")
	asString = strings.TrimSpace(asString)
	safeString := url.QueryEscape(asString)
	safeAsBuffer := bytes.NewBufferString(safeString)
	w.Write(safeAsBuffer.Bytes())
}
