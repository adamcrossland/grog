package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/adamcrossland/grog/mtemplate"
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

	data := mtemplate.NewTemplateData(w, r, loadedNamedQueries, content)
	renderErr := mtemplate.RenderFile(content.Template, w, data)
	if renderErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error rendering post template: %v", renderErr)
		log.Printf("Error rendering post template: %v", renderErr)
	}
}

func putContent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	title := formStringValueOrDef(r, "content_title")
	body := formStringValRequired(w, r, "content_body")
	summary := formStringValueOrDef(r, "content_summary")
	template := formStringValueOrDef(r, "content_template")
	parentID := formIntValueOrDef(r, "content_parent")
	contentID := formIntValueOrDef(r, "content_id")

	if contentID > 0 {
		var getErr error

		oldContent, getErr := grog.GetContent(contentID)
		if getErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error retrieving Content for update; changes not saved")
			log.Printf("error retrieving Content %d for update: %v", contentID, getErr)
			return
		}

		oldContent.UpdateTitle(title)
		oldContent.Summary = summary
		oldContent.Body = body

		saveErr := oldContent.Save()
		if saveErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error updating content: %v", saveErr)
			log.Printf("error updating content: %v", saveErr)
		} else {
			http.Redirect(w, r, urlForContent(*oldContent), http.StatusSeeOther)
		}
	} else {
		newlyAdded := grog.NewContent(title, summary, body, "", template)
		newlyAdded.Parent = parentID

		saveErr := newlyAdded.Save()
		if saveErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error saving new content: %v", saveErr)
			log.Printf("error saving new content: %v", saveErr)
		} else {
			http.Redirect(w, r, urlForContent(*newlyAdded), http.StatusSeeOther)
		}
	}

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
