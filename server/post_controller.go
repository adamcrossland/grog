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
	case "PUT":
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

	title, titleOK := r.Form["title"]
	if !titleOK || len(title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "title must be provided and cannot be empty")
		return
	}

	body, bodyOK := r.Form["body"]
	if !bodyOK || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "body must be provided and cannot be empty")
		return
	}

	summary, _ := r.Form["summary"]

	newlyAdded := grog.NewPost(title[0], summary[0], body[0], "")
	saveErr := newlyAdded.Save()
	if saveErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error saving new post: %v", saveErr)
		log.Printf("error saving new post: %v", saveErr)
	}

	return
}
