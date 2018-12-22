package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/adamcrossland/grog/migrations"

	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/models"
	"github.com/gorilla/mux"
)

var grog *model.GrogModel

func main() {
	// Set up backing database
	dbFilename := os.Getenv("GROG_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable GROG_DATABASE_FILE must be set")
	}

	db := manageddb.NewManagedDB(dbFilename, "sqlite3", migrations.DatabaseMigrations, false)
	grog = model.NewModel(db)

	// Set up request routing
	r := mux.NewRouter()

	//r.HandleFunc("/client", client)
	//r.HandleFunc("/robots.txt", robots)
	r.HandleFunc("/post/{id}", postController)
	//r.HandleFunc("/", client)
	http.Handle("/", r)

	servingAddress := os.Getenv("GROG_SERVER_ADDRESS")
	if servingAddress == "" {
		panic("environment variable GROG_SERVER_ADDRESS must be set")
	}
	fmt.Printf("Listening on %s\n", servingAddress)

	certPath := os.Getenv("GROG_SERVER_CERTPATH")
	if certPath == "" {
		panic("enviornment variable GROG_SERVER_CERTPATH must be set")
	}
	keyPath := os.Getenv("GROG_SERVER_KEYPATH")
	if keyPath == "" {
		panic("enviornment variable GROG_SERVER_KEYPATH must be set")
	}

	//go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	httpErr := http.ListenAndServeTLS(servingAddress, certPath, keyPath, nil)
	if httpErr != nil {
		log.Fatalf("error starting web server: %v\n", httpErr)
	}
}

func postController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		postID, parseErr := strconv.Atoi(vars["id"])
		if parseErr != nil {
			// return 400 error here
		}

		post, postErr := grog.GetPost(int64(postID))
		if postErr != nil {
			log.Printf("Error retrieving post(%d): %v", postID, postErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Title: %s\n\nSummary: %s\n\n%s", post.Title, post.Summary, string(post.Body))
	}

	return
}

func dbFileReader(contentid string) (data []byte, err error) {
	var content *model.Asset
	content, err = grog.GetAsset(contentid)
	if err == nil {
		copy(data, content.Content)
	}

	return
}
