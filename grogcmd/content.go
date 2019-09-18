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

var contentColumns = []string{
	"ID",
	"Title",
	"Summary",
	"Body",
	"Slug",
	"Template",
	"Parent ID",
	"Author ID",
	"Added",
	"Modified",
}

func listContent(longListing bool) {
	allContent, err := grog.AllContents()

	if err != nil {
		fmt.Printf("error loading content: %v\n", err)
		os.Exit(-1)
	}

	var columnData [][]string
	if longListing {
		columnData = make([][]string, len(contentColumns))

		for i, v := range contentColumns {
			columnData[i] = make([]string, 2)
			columnData[i][0] = v
		}
	} else {
		columnData = make([][]string, len(allContent)+1)
		columnData[0] = make([]string, len(contentColumns))
		for i, v := range contentColumns {
			columnData[0][i] = v
		}
	}

	for row := 0; row < len(allContent); row++ {
		contentRow := allContent[row]
		if !longListing {
			columnData[row+1] = make([]string, 10)
		}

		// Make sure that the part of the Body that we show does not have CRLFs in it
		if !longListing {
			contentRow.Body = strings.Replace(contentRow.Body, "\n", "\\n", -1)
		}

		if longListing {
			columnData[0][1] = fmt.Sprintf("%d", contentRow.ID)
			columnData[1][1] = contentRow.Title
			columnData[2][1] = contentRow.Summary
			columnData[3][1] = contentRow.Body
			columnData[4][1] = contentRow.Slug
			columnData[5][1] = contentRow.Template
			columnData[6][1] = fmt.Sprintf("%d", contentRow.Parent)
			columnData[7][1] = fmt.Sprintf("%d", contentRow.Author)
			columnData[8][1] = fmt.Sprintf("%d %d %d", contentRow.Added.Val().Month(),
				contentRow.Added.Val().Day(), contentRow.Added.Val().Year())
			columnData[9][1] = fmt.Sprintf("%d %d %d", contentRow.Modified.Val().Month(),
				contentRow.Modified.Val().Day(), contentRow.Modified.Val().Year())
		} else {
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

		if longListing {
			tabularOutput(columnData)
			fmt.Println() // vertical space between entries
		}
	}

	if !longListing {
		tabularOutput(columnData)
	}
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

func updateContentBody(contentID int64, source *os.File) {
	content, contentErr := grog.GetContent(contentID)
	if contentErr != nil {
		fmt.Printf("error loading content item %d: %v\n", contentID, contentErr)
		return
	}

	newBody := strings.TrimSuffix(readStringToEOL(source), "\n")
	content.Body = newBody

	saveErr := content.Save()
	if saveErr != nil {
		fmt.Printf("error saving content item %d: %v\n", contentID, saveErr)
	}
}

func deleteContent(contentID int64) {
	content, contentErr := grog.GetContent(contentID)
	if contentErr != nil {
		fmt.Printf("error loading content item %d: %v\n", contentID, contentErr)
		return
	}

	delErr := content.Delete()
	if delErr != nil {
		fmt.Printf("error deleting content item %d: %v\n", contentID, delErr)
		return
	}
}
