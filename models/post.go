package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Post struct {
	model   *GrogModel
	ID      int64
	Title   string
	Summary string
	Body    string
	Added   NullTime
	Edited  NullTime
}

func (model *GrogModel) NewPost(title string, summary string, body string) *Post {

	newPost := new(Post)
	newPost.ID = -1 // Not set value
	newPost.Title = title
	newPost.Summary = summary
	newPost.Body = body
	newPost.model = model

	return newPost
}

func (model *GrogModel) GetPost(id int64) (*Post, error) {
	var foundPost *Post
	var err error
	var title string
	var summary string
	var body string
	var added int64
	var edited int64

	row := model.db.DB.QueryRow("select title, summary, body, added, edited from Posts where id = ?", id)

	if row.Scan(&title, &summary, &body, &added, &edited) != sql.ErrNoRows {
		foundPost = model.NewPost(title, summary, body)
		foundPost.ID = id
		foundPost.Title = title
		foundPost.Summary = summary
		foundPost.Body = body
		foundPost.Added.Set(time.Unix(added, 0))
		foundPost.Edited.Set(time.Unix(edited, 0))
	} else {
		err = fmt.Errorf("No post with id %d", id)
	}

	return foundPost, err
}

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

		insertResult, err := post.model.db.DB.Exec("insert into posts (title, summary, body, added, edited) values (?, ?, ?, ?, ?)",
			post.Title, post.Summary, post.Body, post.Added.Unix(), post.Edited.Unix())
		if err == nil {
			post.ID, err = insertResult.LastInsertId()
		}

		saveError = err
	} else {
		// Exists, do update
		_, err := post.model.db.DB.Exec(`update posts set title = ?, summary = ?, body = ?,
				added = ?, edited = ? where Id = ?`, post.Title, post.Summary, post.Body,
			post.Added.Unix(), post.Edited.Unix(), post.ID)
		saveError = err
	}

	return saveError
}

func (post Post) IndexSet() bool {
	return post.ID != -1
}
