package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func loadAsset(rootdir string, asset string, forExternal bool) {
	info, infoError := os.Stat(asset)
	if infoError != nil {
		fmt.Printf("error: %v\n", infoError)
		return
	}

	if info.IsDir() {
		os.Chdir(asset)
		filepath.Walk(".", walkLoader(forExternal))
	} else {
		grog = getModel()
		assetName := strings.Replace(asset, rootdir, "", 1)
		fmt.Printf("loading %s as %s\n", asset, assetName)
	}
}

func walkLoader(forExternal bool) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		curDir, _ := os.Getwd()
		fullPath := filepath.Join(curDir, path)
		fileData, fileErr := ioutil.ReadFile(fullPath)
		if fileErr != nil {
			log.Printf("error reading %s: %v", path, fileErr)
			return fileErr
		}
		fileMimeType := mime.TypeByExtension(filepath.Ext(path))

		fmt.Printf("loading %s as %s\n", path, fileMimeType)
		grog = getModel()

		newAsset := grog.NewAsset(path, fileMimeType)
		_, writeErr := newAsset.Write(fileData)
		if writeErr != nil {
			log.Printf("error copying %s file data to new asset: %v", path, writeErr)
			return writeErr
		}

		newAsset.ServeExternal = forExternal
		saveErr := newAsset.Save()
		if saveErr != nil {
			log.Printf("error saving asset %s to database: %v", path, saveErr)
			return saveErr
		}

		return nil
	}
}

func setAssetProps(assetname string, props []BoolProperty) {
	grog = getModel()

	if grog.AssetExists(assetname) {
		existingAsset, existsErr := grog.GetAsset(assetname)
		if existsErr == nil {
			for _, prop := range props {
				switch prop.Name {
				case "external":
					existingAsset.ServeExternal = prop.Value
				case "render":
					existingAsset.Rendered = prop.Value
				}
			}

			existingAsset.Save()
		} else {
			log.Printf("error accessing asset %s: %v", assetname, existsErr)
		}
	} else {
		log.Printf("asset %s is not in the database\n", assetname)
	}

}

func listAssets() {
	grog = getModel()

	allAssets, err := grog.AllAssets()
	if err != nil {
		fmt.Printf("error loading assets: %v", err)
		return
	}

	for _, eachAsset := range allAssets {
		fmt.Printf("%s\t%s\t", eachAsset.Name, eachAsset.MimeType)
		if eachAsset.ServeExternal {
			fmt.Printf("+ext,")
		} else {
			fmt.Printf("-ext,")
		}

		if eachAsset.Rendered {
			fmt.Printf("+render\t")
		} else {
			fmt.Printf("-render\t")
		}

		fmt.Printf("%v\t%v\n", eachAsset.Added.Time.Format(time.UnixDate),
			eachAsset.Modified.Time.Format(time.UnixDate))
	}
}
