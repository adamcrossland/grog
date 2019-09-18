package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func readDocument(source io.Reader, terminator string) string {
	var readData string

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		readLine := scanner.Text()
		if readLine != terminator {
			if len(readData) > 0 && len(readLine) > 0 {
				readData += "\n"
			}
			readData = readData + readLine
		} else {
			break
		}
	}

	return readData
}

func readStringToEOL(source io.Reader) string {
	reader := bufio.NewReader(source)
	readLine, _ := reader.ReadString('\n')

	return readLine
}

func readIntToEOL(source io.Reader) (int64, bool) {
	asText := strings.TrimSuffix(readStringToEOL(source), "\n")
	if len(asText) > 0 {
		asInt, err := strconv.ParseInt(asText, 10, 64)
		if err == nil {
			return asInt, true
		}
	}

	return 0, false
}
