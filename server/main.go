package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/adamcrossland/grog/migrations"

	"bitbucket.org/adamcrossland/mtemplate"
	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/models"
	"github.com/gorilla/mux"
)

var grog *model.GrogModel
var noCaching bool = false

func main() {
	argsWithoutProg := os.Args[1:]
	for i := 0; i < len(argsWithoutProg); i++ {
		switch strings.ToLower(argsWithoutProg[i]) {
		case "--no-cache":
			noCaching = true
		}
	}

	// Set up backing database
	dbFilename := os.Getenv("GROG_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable GROG_DATABASE_FILE must be set")
	}

	db := manageddb.NewManagedDB(dbFilename, "sqlite3", migrations.DatabaseMigrations, false)
	grog = model.NewModel(db)

	// Set up templating engine to read files fromthe database
	mtemplate.TemplateSourceReader = dbFileReader

	// Set up request routing
	r := mux.NewRouter()

	r.HandleFunc("/post/{id}", postController)
	r.HandleFunc("/asset/{id}", assetController)
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

	go http.ListenAndServe(":8081", http.HandlerFunc(redirect))

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
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "post %s does not exist", postID)
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

func assetController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		assetID := vars["id"]

		asset, assetErr := grog.GetAsset(assetID)
		if assetErr != nil {
			log.Printf("Error retrieving asset(%s): %v", assetID, assetErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if asset == nil || asset.ServeExternal == false {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-type", asset.MimeType)
		w.Header().Set("Content-length", strconv.Itoa(len(asset.Content)))

		switch asset.MimeType {
		case "text/css", "text/html", "text/plain":
			fmt.Fprintf(w, "%s", string(asset.Content))
		default:
			w.Write(asset.Content)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}

func dbFileReader(contentid string) (data []byte, err error) {
	var content *model.Asset
	content, err = grog.GetAsset(contentid)
	if err == nil {
		data = make([]byte, len(content.Content))
		copy(data, content.Content)
	} else {
		log.Printf("Error while retrieving asset %s: %v\n", contentid, err)
	}

	return
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
