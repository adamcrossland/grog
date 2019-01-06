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
