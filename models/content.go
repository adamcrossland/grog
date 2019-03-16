package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Content models an individual unit of blog content
type Content struct {
	model    *GrogModel
	ID       int64
	Title    string
	Summary  string
	Body     string
	Slug     string
	Template string
	Parent   int64
	Author   int64
	Added    NullTime
	Edited   NullTime
	Children []*Content
}

// NewContent creates a new Content object
func (model *GrogModel) NewContent(title string, summary string, body string, slug string,
	template string) *Content {

	newContent := new(Content)
	newContent.ID = -1 // Not set value
	newContent.Author = 0
	newContent.Title = title
	if len(slug) == 0 && len(newContent.Title) > 0 {
		newContent.Slug = MakeSlug(newContent.Title)
	}
	newContent.Summary = summary
	newContent.Body = body
	newContent.Template = template
	newContent.model = model

	return newContent
}

// GetContent retrieves the Content object identified by the given id from the database
func (model *GrogModel) GetContent(id int64) (*Content, error) {
	var foundContent *Content
	var err error

	contentRow, queryErr := model.db.DB.Query(`select id, title, summary, body, slug, template,
												parent, author, added, edited from Content where id = ?`, id)

	if queryErr == nil {
		defer contentRow.Close()

		foundContent = model.readContentFromRow(contentRow)

		if foundContent == nil {
			err = fmt.Errorf("no Content with id %d", id)
		}
	} else {
		err = fmt.Errorf("database error while reading Content: %v", queryErr)
	}

	return foundContent, err
}

// GetContentBySlug returns the Content object with the given slug
func (model *GrogModel) GetContentBySlug(slugged string) (*Content, error) {
	var foundContent *Content
	var err error

	contentRow, queryErr := model.db.DB.Query(`select id, title, summary, body, slug, template,
		parent, author, added, edited from Content where slug = ?`, slugged)

	if queryErr == nil {
		defer contentRow.Close()

		foundContent = model.readContentFromRow(contentRow)

		if foundContent == nil {
			err = fmt.Errorf("no Content with slug %s", slugged)
		}
	} else {
		err = fmt.Errorf("database error while reading Content: %v", queryErr)
	}

	return foundContent, err
}

func (model *GrogModel) readContentFromRow(rows *sql.Rows) *Content {
	var foundContent *Content

	if rows.Next() {
		var id int64
		var title string
		var summary string
		var body string
		var slug string
		var template string
		var parent int64
		var author int64
		var added int64
		var edited int64

		if rows.Scan(&id, &title, &summary, &body, &slug, &template, &parent, &author, &added, &edited) != sql.ErrNoRows {
			foundContent = model.NewContent(title, summary, body, slug, template)
			foundContent.ID = id
			foundContent.Parent = parent
			foundContent.Author = author
			foundContent.Added.Set(time.Unix(added, 0))
			foundContent.Edited.Set(time.Unix(edited, 0))
		}
	}

	return foundContent
}

// Save writes the Content object to the database
func (content *Content) Save() error {
	var saveError error

	if content.ID == -1 {
		// New, do insert
		if content.Added.IsNull() {
			content.Added.Set(time.Now())
		}

		if content.Edited.IsNull() {
			content.Edited.Set(content.Added.Time)
		}

		insertResult, err := content.model.db.DB.Exec(`insert into content (title, summary, body, slug, 
			template, parent, author, added, edited) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			content.Title, content.Summary, content.Body, content.Slug, content.Template,
			content.Parent, content.Author, content.Added.Unix(), content.Edited.Unix())
		if err == nil {
			content.ID, err = insertResult.LastInsertId()
			if err != nil {
				fmt.Printf("err getting id for newly-inserted Content: %v", err)
			}
		}

		saveError = err
	} else {
		content.Edited.Set(time.Now())
		// Exists, do update
		_, err := content.model.db.DB.Exec(`update content set title = ?, summary = ?, body = ?, slug = ?,
				template = ?, parent = ?, author = ?, added = ?, edited = ? where Id = ?`, content.Title,
			content.Summary, content.Body, content.Slug, content.Template, content.Parent,
			content.Author, content.Added.Unix(), content.Edited.Unix(), content.ID)
		saveError = err
	}

	return saveError
}

// IndexSet return true if the Content object has an ID set rather than the default value
func (content Content) IndexSet() bool {
	return content.ID != -1
}

// IncludeChildren populates the Content's Children property with all of the Content
// objects in the database whose Parent field is equal to the Content's ID.
func (content *Content) IncludeChildren() *Content {

	var foundContent *Content

	contentRows, queryErr := content.model.db.DB.Query(`select id, title, summary, body, slug, template,
	parent, author, added, edited from Content where parent = ?`, content.ID)

	if queryErr == nil {
		defer contentRows.Close()

		foundContent = content.model.readContentFromRow(contentRows)
		for foundContent != nil {
			fmt.Printf("found child content: %+v\n", foundContent)
			if content.Children == nil {
				content.Children = make([]*Content, 1, 10)
				content.Children[0] = foundContent
			} else {
				content.Children = append(content.Children, foundContent)
			}
			foundContent = content.model.readContentFromRow(contentRows)
		}
	}

	return content
}

// MakeSlug creates a URL-safe version of a string, usually the Title of a Content.
func MakeSlug(toSlug string) string {
	a := strings.ToLower(toSlug)
	b := make([]rune, 0)
	prevSpace := false
	for _, rune := range a {
		if unicode.IsSpace(rune) {
			if !prevSpace {
				b = append(b, '-')
				prevSpace = true
			}
		} else {
			prevSpace = false
			if unicode.IsLower(rune) {
				b = append(b, rune)
			}
		}
	}

	return string(b)
}
