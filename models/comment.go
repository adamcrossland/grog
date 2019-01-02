package model

import (
	"fmt"
	"time"
)

// Comment contains all of the information associated with a comment
// on a Post.
type Comment struct {
	model   *GrogModel
	ID      int64
	Content string
	Author  int64
	Added   NullTime
	Post    int64
}

func (grog *GrogModel) NewComment(id int64, content string, author int64, post int64, added NullTime) *Comment {
	newComment := new(Comment)
	newComment.ID = id
	newComment.Content = content
	newComment.Author = author
	newComment.Post = post
	newComment.Added = added

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

// Comments retrieves all comments related to the Post.
func (post *Post) Comments() ([]*Comment, error) {

	rows, err := post.model.db.DB.Query("select id, content, author, post, added from Comments where post = ? order by added", post.ID)
	if err != nil {
		return nil, fmt.Errorf("error reading comments for post %d: %v", post.ID, err)
	}

	defer rows.Close()

	foundComments := make([]*Comment, 0)

	for rows.Next() {
		var id, author, fromPost, added int64
		var content string

		err = rows.Scan(&id, &content, &author, &fromPost, &added)
		if err != nil {
			return nil, fmt.Errorf("error scanning comments query result for post %d: %v", post.ID, err)
		}

		var addedTime NullTime
		addedTime.Set(time.Unix(added, 0))
		newComment := post.model.NewComment(id, content, author, fromPost, addedTime)
		foundComments = append(foundComments, newComment)
	}

	return foundComments, nil
}
