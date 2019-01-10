package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/adamcrossland/mtemplate"
	model "github.com/adamcrossland/grog/models"
	"github.com/gorilla/mux"
)

func postController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		postID, ok := vars["id"]
		if !ok || len(postID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "post id or slug must be provided")
		}

		getPost(w, r, postID)
	case "PUT", "POST":
		putPost(w, r)
		// TODO: This must be authenticated and authorized

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}

func getPost(w http.ResponseWriter, r *http.Request, postID string) {

	var post *model.Post
	var postErr error

	parsedID, parseErr := strconv.Atoi(postID)

	if parseErr == nil {
		// Numeric argument was provided, so retrieve the post with the ID
		post, postErr = grog.GetPost(int64(parsedID))
	} else {
		// Non-numeric argument, treat it as a slug
		post, postErr = grog.GetPostBySlug(postID)
	}

	if postErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving post %s: %v", postID, postErr)

		return
	}

	if post == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Could not find post %s", postID)
		return
	}

	post.LoadComments()

	if noCaching {
		mtemplate.ClearFromCache("post.html")
		mtemplate.ClearFromCache("base.html")
	}

	renderErr := mtemplate.RenderFile("post.html", w, post)
	if renderErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error rendering post template: %v", renderErr)
		log.Printf("Error rendering post template: %v", renderErr)
	}
}

func putPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	title, titleOK := r.Form["post_title"]
	if !titleOK || len(title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "title must be provided and cannot be empty")
		return
	}

	body, bodyOK := r.Form["post_content"]
	if !bodyOK || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "content must be provided and cannot be empty")
		return
	}

	summary, _ := r.Form["post_summary"]

	var newlyAdded *model.Post

	postID, postIDOK := r.Form["post_id"]

	if postIDOK && len(postID[0]) > 0 {
		var getErr error
		intID, _ := strconv.Atoi(postID[0])
		newlyAdded, getErr = grog.GetPost(int64(intID))
		if getErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error retrieving Post for update; changes not saved")
			log.Printf("error retrieving Post %d for update: %v", postID[0], getErr)
			return
		}
		if newlyAdded.Title != title[0] {
			newlyAdded.Title = model.MakeSlug(title[0])
		}
		newlyAdded.Title = title[0]
		newlyAdded.Summary = summary[0]
		newlyAdded.Body = body[0]
	} else {
		newlyAdded = grog.NewPost(title[0], summary[0], body[0], "")
	}

	saveErr := newlyAdded.Save()
	if saveErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error saving new post: %v", saveErr)
		log.Printf("error saving new post: %v", saveErr)
	}

	http.Redirect(w, r, urlForPost(*newlyAdded), http.StatusSeeOther)

	return
}

func postEditor(w http.ResponseWriter, r *http.Request) {
	var onPost *model.Post

	r.ParseForm()
	id, idOK := r.Form["id"]
	if idOK {
		// We are being asked to edit an existing post. Retrieve it and pass it to the template
		parsedID, parseErr := strconv.Atoi(id[0])
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "could not parse %s as a valid post id, which must be numeric", id)
			return
		}

		var getPostErr error
		onPost, getPostErr = grog.GetPost(int64(parsedID))
		if getPostErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error retrieving post; try again later")
			log.Printf("error retrieving post %d: %v", parsedID, getPostErr)
			return
		}
	}

	if noCaching {
		mtemplate.ClearFromCache("newpost.html")
	}

	mtemplate.RenderFile("newpost.html", w, onPost)
}

func urlForPost(post model.Post) string {
	return "/post/" + post.Slug
}

func postCommentController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id of post must be provided")
		return
	}

	switch r.Method {
	case "GET":
		getComment(w, r, postID)
	case "POST":
		postComment(w, r, postID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getComment(w http.ResponseWriter, r *http.Request, postID string) {
	// Show the Post page with an open comment area. This will generally only
	// happen when JS is not available in the browser.
	var post *model.Post
	var postErr error

	parsedID, parseErr := strconv.Atoi(postID)

	if parseErr == nil {
		// Numeric argument was provided, so retrieve the post with the ID
		post, postErr = grog.GetPost(int64(parsedID))
	} else {
		// Non-numeric argument, treat it as a slug
		post, postErr = grog.GetPostBySlug(postID)
	}

	if postErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving post %s: %v", postID, postErr)

		return
	}

	if post == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Could not find post %s", postID)
		return
	}

	post.LoadComments()

	if noCaching {
		mtemplate.ClearFromCache("new_comment.html")
		mtemplate.ClearFromCache("post.html")
		mtemplate.ClearFromCache("base.html")
	}

	renderErr := mtemplate.RenderFile("new_comment.html", w, post)
	if renderErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error rendering post template: %v", renderErr)
		log.Printf("Error rendering post template: %v", renderErr)
	}
}

func postComment(w http.ResponseWriter, r *http.Request, postID string) {
	r.ParseForm()

	content, contentOK := r.Form["comment"]
	if !contentOK || len(content) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "content must be provided and cannot be empty")
		return
	}

	parsedID, parseErr := strconv.Atoi(postID)

	if parseErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "provided post id %s could not be converted to a post id", postID)
		return
	}

	onPost, onPostErr := grog.GetPost(int64(parsedID))
	if onPostErr != nil {
		log.Printf("error retrieving post %d: %v", parsedID, onPostErr)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error retrieving post. try again later.")
		return
	}

	// TODO: once authentication has been added; this code will need to look up
	// the actual user. For now, just use user 1.
	commentingUser, commentingUserErr := grog.GetUser(1)
	if commentingUserErr != nil {
		log.Printf("could not load user %d: %v", 1, commentingUserErr)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "could not load user. try again later.")
		return
	}

	_, newCommentErr := onPost.AddComment(content[0], *commentingUser)
	if newCommentErr != nil {
		log.Printf("error adding new comment to post %d: %v", parsedID, newCommentErr)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error saving comment. try again later.")
		return
	}

	// Comment was saved successfully. Redirect nack to the post itself, so the user
	// can see their post in context.
	http.Redirect(w, r, urlForPost(*onPost), http.StatusSeeOther)

	return
}
