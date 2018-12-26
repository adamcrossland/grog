package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/adamcrossland/mtemplate"
	"github.com/gorilla/mux"
)

func postController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		postID, parseErr := strconv.Atoi(vars["id"])
		if parseErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "post %s does not exist", vars["id"])
		}

		post, postErr := grog.GetPost(int64(postID))
		if postErr != nil {
			log.Printf("Error retrieving post(%d): %v", postID, postErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if post == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if noCaching {
			mtemplate.ClearFromCache("post.html")
			mtemplate.ClearFromCache("base.html")
		}

		renderErr := mtemplate.RenderFile("post.html", w, post)
		if renderErr != nil {
			log.Printf("Error rendering post template: %v", renderErr)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}
