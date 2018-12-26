package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
