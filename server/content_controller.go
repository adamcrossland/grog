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

func contentController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		contentID, ok := vars["id"]
		if !ok || len(contentID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "content id or slug must be provided")
		}

		getContent(w, r, contentID)
	case "PUT", "POST":
		putContent(w, r)
		// TODO: This must be authenticated and authorized

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}

func getContent(w http.ResponseWriter, r *http.Request, contentID string) {

	var content *model.Content
	var contentErr error

	parsedID, parseErr := strconv.Atoi(contentID)

	if parseErr == nil {
		// Numeric argument was provided, so retrieve the post with the ID
		content, contentErr = grog.GetContent(int64(parsedID))
	} else {
		// Non-numeric argument, treat it as a slug
		content, contentErr = grog.GetContentBySlug(contentID)
	}

	if contentErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error retrieving post %s: %v", contentID, contentErr)

		return
	}

	if content == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Could not find post %s", contentID)
		return
	}

	content.IncludeChildren()

	renderErr := mtemplate.RenderFile(content.Template, w, content)
	if renderErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error rendering post template: %v", renderErr)
		log.Printf("Error rendering post template: %v", renderErr)
	}
}

func putContent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	title, titleOK := r.Form["content_title"]
	if !titleOK || len(title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "title must be provided and cannot be empty")
		return
	}

	body, bodyOK := r.Form["content_body"]
	if !bodyOK || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "content must be provided and cannot be empty")
		return
	}

	summary, _ := r.Form["content_summary"]
	template, _ := r.Form["content_template"]

	var newlyAdded *model.Content

	contentID, contentIDOK := r.Form["content_id"]

	if contentIDOK && len(contentID[0]) > 0 {
		var getErr error
		intID, _ := strconv.Atoi(contentID[0])
		newlyAdded, getErr = grog.GetContent(int64(intID))
		if getErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error retrieving Post for update; changes not saved")
			log.Printf("error retrieving Post %s for update: %v", contentID[0], getErr)
			return
		}
		if len(newlyAdded.Title) > 0 && newlyAdded.Title != title[0] {
			newlyAdded.Slug = model.MakeSlug(title[0])
		}
		newlyAdded.Title = title[0]
		newlyAdded.Summary = summary[0]
		newlyAdded.Body = body[0]
	} else {
		var newSlug string

		if len(title[0]) > 0 {
			newSlug = model.MakeSlug(title[0])
		}

		newlyAdded = grog.NewContent(title[0], summary[0], body[0], newSlug, template[0])
	}

	saveErr := newlyAdded.Save()
	if saveErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error saving new content: %v", saveErr)
		log.Printf("error saving new content: %v", saveErr)
	}

	http.Redirect(w, r, urlForContent(*newlyAdded), http.StatusSeeOther)

	return
}

func urlForContent(content model.Content) string {
	var url string
	if len(content.Slug) > 0 {
		url = "/content/" + content.Slug
	} else {
		url = "/content/" + string(content.ID)
	}

	return url
}
