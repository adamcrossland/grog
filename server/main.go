package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/adamcrossland/grog/migrations"

	"github.com/adamcrossland/grog/manageddb"
	model "github.com/adamcrossland/grog/models"
	"github.com/adamcrossland/grog/mtemplate"
	"github.com/gorilla/mux"
)

var grog *model.GrogModel
var loadedNamedQueries map[string]model.NamedQueryFunc

func main() {
	argsWithoutProg := os.Args[1:]
	for i := 0; i < len(argsWithoutProg); i++ {
		switch strings.ToLower(argsWithoutProg[i]) {
		case "--no-cache":
			mtemplate.Cache = false
		}
	}

	// Set up backing database
	dbFilename := os.Getenv("GROG_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable GROG_DATABASE_FILE must be set")
	}

	db := manageddb.NewManagedDB(dbFilename, "sqlite3", migrations.DatabaseMigrations, false)
	grog = model.NewModel(db)

	// Load namedqueries
	loadedNamedQueries = grog.LoadNamedQueries()

	// Set up templating engine to read files from the database
	mtemplate.TemplateSourceReader = dbFileReader

	mtemplate.CustomFormatters = mtemplate.FormatterMap{
		"shortdate": ShortDateFormatter,
		"trunc":     TruncFormatter,
	}

	// Set up request routing
	r := mux.NewRouter()

	r.HandleFunc("/content/{id:[a-zA-z0-9/\\-_\\.]+}", contentController)
	r.HandleFunc("/content", contentController)
	r.HandleFunc("/asset/{id:[a-zA-Z0-9/\\-_\\.]+}", assetController)
	r.HandleFunc("/asset", assetController)
	r.HandleFunc("/{id:[a-zA-Z0-9/\\-_\\.]+}", assetController)
	r.HandleFunc("/", assetController)
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

func dbFileReader(assetID string) (data []byte, err error) {
	var asset *model.Asset
	asset, err = grog.GetAsset(assetID)
	if err == nil {
		data = asset.Content
	} else {
		log.Printf("Error while retrieving asset %s: %v\n", assetID, err)
	}

	return
}
