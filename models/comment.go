package model

import (
	"fmt"
	"time"
)

// Comment contains all of the information associated with a comment
// on a Post.
type Comment struct {
	model       *GrogModel
	ID          int64
	Content     string
	Author      int64
	Added       NullTime
	Post        int64
	AuthorName  string
	AuthorEmail string
}

// NewComment creates a new Comment instance
func (model *GrogModel) NewComment(id int64, content string, author int64, post int64, added NullTime) *Comment {
	newComment := new(Comment)
	newComment.ID = id
	newComment.Content = content
	newComment.Author = author
	newComment.Post = post
	newComment.Added = added
	newComment.model = model

	return newComment
}

// AddComment adds a new Comment to this Post
func (post *Post) AddComment(content string, author User) (*Comment, error) {
	var curTime NullTime
	curTime.Set(time.Now())
	newComment := post.model.NewComment(-1, content, author.ID, post.ID, curTime)

	insertResult, err := post.model.db.DB.Exec("insert into comments (Content, Author, Post, added) values (?, ?, ?, ?)",
		content, author.ID, post.ID, newComment.Added.Unix())
	if err == nil {
		newComment.ID, err = insertResult.LastInsertId()
		return newComment, nil
	}

	return nil, err
}

// LoadComments retrieves all comments related to the Post. They are returned by this call, but are
// also available in the Post's Comments property.
func (post *Post) LoadComments() ([]*Comment, error) {

	rows, err := post.model.db.DB.Query(`
		select comments.ID, Content, Author, comments.Added, comments.post, users.email, users.Name
		from Comments inner join users on comments.Author = users.ID where Post=?
		order by comments.added`, post.ID)
	if err != nil {
		return nil, fmt.Errorf("error reading comments for post %d: %v", post.ID, err)
	}

	defer rows.Close()

	foundComments := make([]*Comment, 0)

	for rows.Next() {
		var id, author, fromPost, added int64
		var content, authorName, authorEmail string

		err = rows.Scan(&id, &content, &author, &added, &fromPost, &authorEmail, &authorName)
		if err != nil {
			return nil, fmt.Errorf("error scanning comments query result for post %d: %v", post.ID, err)
		}

		var addedTime NullTime
		addedTime.Set(time.Unix(added, 0))
		newComment := post.model.NewComment(id, content, author, fromPost, addedTime)
		newComment.AuthorEmail = authorEmail
		newComment.AuthorName = authorName

		foundComments = append(foundComments, newComment)
	}

	post.Comments = foundComments
	return foundComments, nil
}
