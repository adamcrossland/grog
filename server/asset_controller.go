package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/adamcrossland/grog/mtemplate"
	"github.com/gorilla/mux"
)

func assetController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "GET":
		assetID := vars["id"]

		if assetID == "" {
			assetID = "index"
		}

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

		mimeTypeParts := strings.Split(asset.MimeType, ";")
		mimeTypeAbbrv := mimeTypeParts[0]

		switch mimeTypeAbbrv {
		case "text/css", "text/html", "text/plain", "text/javascript":
			if asset.Rendered {
				parsedTemplate := mtemplate.MustParse(string(asset.Content), nil)
				tdata := mtemplate.NewTemplateData(w, r, loadedNamedQueries, nil)
				parsedTemplate.Execute(w, tdata)
			} else {
				fmt.Fprintf(w, "%s", string(asset.Content))
			}
		default:
			w.Write(asset.Content)
		}
	case "PUT", "POST":
		r.ParseForm()

		assetName, assetNameOK := r.Form["name"]
		if !assetNameOK || len(assetName) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "must include 'name' paramter")
			return
		}

		assetType, assetTypeOK := r.Form["mimetype"]
		if !assetTypeOK || len(assetType) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "must include 'mimetype' paramter")
			return
		}

		assetExternal, assetExternalOK := r.Form["external"]
		saveExternal := false
		if !assetExternalOK || len(assetExternal) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "must include 'external' paramter")
			return
		}
		switch assetExternal[0] {
		case "y", "Y", "yes", "YES", "Yes", "t", "T", "true", "TRUE", "True":
			saveExternal = true
		case "n", "N", "no", "NO", "No", "f", "F", "false", "FALSE", "False":
			saveExternal = false
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Parameter 'external' has unrecognized value")
			return
		}

		assetContent, assetContentOK := r.Form["content"]
		if !assetContentOK || len(assetContent) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "must include 'content' paramter")
			return
		}
		decodedContent, decodeErr := base64.StdEncoding.DecodeString(assetContent[0])
		if decodeErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error decoding content. not base64 encoded?")
		}

		asset := grog.NewAsset(assetName[0], assetType[0])
		asset.ServeExternal = saveExternal
		asset.Write(decodedContent)
		assetSaveErr := asset.Save()
		if assetSaveErr != nil {
			log.Printf("error saving asset: %v\n", assetSaveErr)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error saving content; content not added")
		}
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}
