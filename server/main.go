package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/adamcrossland/grog/migrations"

	"bitbucket.org/adamcrossland/mtemplate"
	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/models"
	"github.com/gorilla/mux"
)

var grog *model.GrogModel
var noCaching = false

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
	r.HandleFunc("/post", postController)
	r.HandleFunc("/asset/{id}", assetController)
	r.HandleFunc("/asset", assetController)
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

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
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
