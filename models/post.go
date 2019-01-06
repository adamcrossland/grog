package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Post models an individual unit of blog content
type Post struct {
	model    *GrogModel
	ID       int64
	Title    string
	Summary  string
	Body     string
	Slug     string
	Added    NullTime
	Edited   NullTime
	Comments []*Comment
}

// NewPost creates a new Post object
func (model *GrogModel) NewPost(title string, summary string, body string, slug string) *Post {

	newPost := new(Post)
	newPost.ID = -1 // Not set value
	newPost.Title = title
	if len(slug) == 0 {
		newPost.Slug = MakeSlug(newPost.Title)
	}
	newPost.Summary = summary
	newPost.Body = body
	newPost.model = model

	return newPost
}

// GetPost retrieves the Post object identified by the given id from the database
func (model *GrogModel) GetPost(id int64) (*Post, error) {
	var foundPost *Post
	var err error
	var title string
	var summary string
	var body string
	var slug string
	var added int64
	var edited int64

	row := model.db.DB.QueryRow("select title, summary, body, slug, added, edited from Posts where id = ?", id)

	if row.Scan(&title, &summary, &body, &slug, &added, &edited) != sql.ErrNoRows {
		foundPost = model.NewPost(title, summary, body, slug)
		foundPost.ID = id
		foundPost.Title = title
		foundPost.Summary = summary
		foundPost.Body = body
		foundPost.Slug = slug
		foundPost.Added.Set(time.Unix(added, 0))
		foundPost.Edited.Set(time.Unix(edited, 0))
	} else {
		err = fmt.Errorf("No post with id %d", id)
	}

	return foundPost, err
}

// GetPostBySlug returns the Post object with the given slug
func (model *GrogModel) GetPostBySlug(slugged string) (*Post, error) {
	var foundPost *Post
	var err error
	var id int64
	var title string
	var summary string
	var body string
	var slug string
	var added int64
	var edited int64

	row := model.db.DB.QueryRow("select id, title, summary, body, slug, added, edited from Posts where slug = ?", slugged)

	if row.Scan(&id, &title, &summary, &body, &slug, &added, &edited) != sql.ErrNoRows {
		foundPost = model.NewPost(title, summary, body, slug)
		foundPost.ID = id
		foundPost.Title = title
		foundPost.Summary = summary
		foundPost.Body = body
		foundPost.Slug = slug
		foundPost.Added.Set(time.Unix(added, 0))
		foundPost.Edited.Set(time.Unix(edited, 0))
	} else {
		err = fmt.Errorf("No post with slug %s", slugged)
	}

	return foundPost, err
}

// Save writes the Post object to the database
func (post *Post) Save() error {
	var saveError error

	if post.ID == -1 {
		// New, do insert
		if post.Added.IsNull() {
			post.Added.Set(time.Now())
		}

		if post.Edited.IsNull() {
			post.Edited.Set(post.Added.Time)
		}

		insertResult, err := post.model.db.DB.Exec("insert into posts (title, summary, body, slug, added, edited) values (?, ?, ?, ?, ?, ?)",
			post.Title, post.Summary, post.Body, post.Slug, post.Added.Unix(), post.Edited.Unix())
		if err == nil {
			post.ID, err = insertResult.LastInsertId()
		}

		saveError = err
	} else {
		post.Edited.Set(time.Now())
		// Exists, do update
		_, err := post.model.db.DB.Exec(`update posts set title = ?, summary = ?, body = ?, slug = ?,
				added = ?, edited = ? where Id = ?`, post.Title, post.Summary, post.Body, post.Slug,
			post.Added.Unix(), post.Edited.Unix(), post.ID)
		saveError = err
	}

	return saveError
}

// IndexSet return true if the Post object has an ID set rather than the default value
func (post Post) IndexSet() bool {
	return post.ID != -1
}

// MakeSlug creates a URL-safe version of a string, usually the Title of a Post.
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
