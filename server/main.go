package main

import (
	"os"
	"github.com/adamcrossland/grog/manageddb"
)

func main() {
	// Set up backing database
	dbFilename := os.Getenv("GROG_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable GROG_DATABASE_FILE must be set")
	}

	db := manageddb.NewManagedDB(dbFilename, "sqlite3", databaseMigrations, false)

	// Set up request routing
	r := mux.NewRouter()

	r.HandleFunc("/client", client)
	r.HandleFunc("/robots.txt", robots)
	r.HandleFunc("/post/{id}", showPost)
	r.HandleFunc("/", client)
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

	go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	httpErr := http.ListenAndServeTLS(servingAddress, certPath, keyPath, nil)
	if httpErr != nil {
		log.Fatalf("error starting web server: %v\n", httpErr)
	}
}

func showPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId, parseErr := strconv.Atoi(vars["id"])

	if parseErr != nil {

	}

}

func dbFileReader(contentid string) (byte[], error) {
	
}