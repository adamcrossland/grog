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
