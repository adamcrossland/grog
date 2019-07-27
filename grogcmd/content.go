package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	model "github.com/adamcrossland/grog/models"
)

func loadContent(title string, summary string, body string, template string) *model.Content {
	newContent := grog.NewContent(title, summary, body, "", template)

	saveErr := newContent.Save()
	if saveErr != nil {
		fmt.Printf("error saving new content: %v\n", saveErr)
		os.Exit(-1)
	}

	return newContent
}

func listContent() {
	allContent, err := grog.AllContents()

	if err != nil {
		fmt.Printf("error loading content: %v\n", err)
		os.Exit(-1)
	}

	columnData := make([][]string, len(allContent)+1)
	columnData[0] = make([]string, 10)
	columnData[0][0] = "ID"
	columnData[0][1] = "Title"
	columnData[0][2] = "Summary"
	columnData[0][3] = "Body"
	columnData[0][4] = "Slug"
	columnData[0][5] = "Template"
	columnData[0][6] = "Parent ID"
	columnData[0][7] = "Author ID"
	columnData[0][8] = "Added"
	columnData[0][9] = "Modified"

	for row := 0; row < len(allContent); row++ {
		contentRow := allContent[row]
		columnData[row+1] = make([]string, 10)

		// Make sure that the part of the Body that we show does not have CRLFs in it
		contentRow.Body = strings.Replace(contentRow.Body, "\n", "\\n", -1)

		columnData[row+1][0] = fmt.Sprintf("%d", contentRow.ID)
		columnData[row+1][1] = fmt.Sprintf("%.10s", contentRow.Title)
		columnData[row+1][2] = fmt.Sprintf("%.10s", contentRow.Summary)
		columnData[row+1][3] = fmt.Sprintf("%.10s", contentRow.Body)
		columnData[row+1][4] = fmt.Sprintf("%.10s", contentRow.Slug)
		columnData[row+1][5] = contentRow.Template
		columnData[row+1][6] = fmt.Sprintf("%d", contentRow.Parent)
		columnData[row+1][7] = fmt.Sprintf("%d", contentRow.Author)

		columnData[row+1][8] = fmt.Sprintf("%d %d %d", contentRow.Added.Val().Month(),
			contentRow.Added.Val().Day(), contentRow.Added.Val().Year())
		columnData[row+1][9] = fmt.Sprintf("%d %d %d", contentRow.Modified.Val().Month(),
			contentRow.Modified.Val().Day(), contentRow.Modified.Val().Year())
	}

	tabularOutput(columnData)
}

func addContent(source io.Reader) {
	fmt.Print("Title: ")
	title := strings.TrimSuffix(readStringToEOL(source), "\n")

	fmt.Print("Summary: ")
	summary := strings.TrimSuffix(readStringToEOL(source), "\n")

	fmt.Println("Body: (__EOF__ to finish)")
	body := strings.TrimSuffix(readDocument(source, "__EOF__"), "\n")

	fmt.Print("Template: ")
	template := strings.TrimSuffix(readStringToEOL(source), "\n")

	fmt.Print("Parent ID: ")
	parentID, parentIDOK := readIntToEOL(source)

	fmt.Printf("Author ID: ")
	authorID, authorIDOK := readIntToEOL(source)

	newContent := grog.NewContent(title, summary, body, "", template)
	if parentIDOK {
		newContent.Parent = parentID
	}
	if authorIDOK {
		newContent.Author = authorID
	}

	newContent.Save()

	fmt.Printf("Added new content with id %v\n", newContent.ID)
}
