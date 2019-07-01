package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
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

	columnData := make([][]string, len(allAssets))

	for row := 0; row < len(allAssets); row++ {
		columnData[row] = make([]string, 6)
		columnData[row][0] = allAssets[row].Name
		columnData[row][1] = allAssets[row].MimeType

		if allAssets[row].ServeExternal {
			columnData[row][2] = "+ext"
		} else {
			columnData[row][2] = "-ext"
		}

		if allAssets[row].Rendered {
			columnData[row][3] = "+rnd"
		} else {
			columnData[row][3] = "-rnd"
		}

		columnData[row][4] = fmt.Sprintf("%d %d %d", allAssets[row].Added.Val().Month(),
			allAssets[row].Added.Val().Day(), allAssets[row].Added.Val().Year())
		columnData[row][5] = fmt.Sprintf("%d %d %d", allAssets[row].Modified.Val().Month(),
			allAssets[row].Modified.Val().Day(), allAssets[row].Modified.Val().Year())
	}

	tabularOutput(columnData)
}
