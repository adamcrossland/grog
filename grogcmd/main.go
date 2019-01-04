package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/migrations"
	model "github.com/adamcrossland/grog/models"
)

var grog *model.GrogModel

func main() {
	args := os.Args

	if len(args) == 1 {
		help()
		os.Exit(-1)
	}

	switch strings.ToLower(args[1]) {
	case "asset":
		if len(args) >= 4 {
			switch strings.ToLower(args[2]) {
			case "add":
				var assetName string
				loadForExternal := false

				if args[3][0] == '-' {
					switch strings.ToLower(args[3]) {
					case "-ext":
						loadForExternal = true
					default:
						fmt.Printf("flag %s not understood\n", args[3])
						helpAssetCmd()
						os.Exit(-1)
					}
					assetName = args[4]
				} else {
					assetName = args[3]
				}

				curDir, _ := os.Getwd()
				loadAsset(curDir, assetName, loadForExternal)
			default:
				fmt.Printf("asset sub-command %s not understood\n", args[2])
				helpAssetCmd()
			}
		} else {
			helpAssetCmd()
			os.Exit(-1)
		}
	default:
		help()
	}
}

func getModel() *model.GrogModel {
	if grog == nil {
		// Set up backing database
		dbFilename := os.Getenv("GROG_DATABASE_FILE")
		if dbFilename == "" {
			panic("environment variable GROG_DATABASE_FILE must be set")
		}

		db := manageddb.NewManagedDB(dbFilename, "sqlite3", migrations.DatabaseMigrations, false)
		grog = model.NewModel(db)
	}

	return grog
}

func help() {
	fmt.Println("Usage:")
	helpAssetCmd()
}

func helpAssetCmd() {
	fmt.Println("\tgrogcmd asset add [-ext] <file|directory>")
}
