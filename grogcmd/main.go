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

func init() {
	grog = getModel()
}

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
			case "mv":
				if len(args) == 5 {
					moveFrom := args[3]
					moveTo := args[4]

					if grog.AssetExists(moveTo) {
						fmt.Printf("asset %s exists; delete it first if you want to give another asset that name\n", moveTo)
						os.Exit(-1)
					}

					fromAsset, getFromAssetErr := grog.GetAsset(moveFrom)
					if getFromAssetErr != nil {
						fmt.Printf("asset %s could not be retrieved, so cannot be renamed\n", moveFrom)
						os.Exit(-1)
					}

					renameErr := fromAsset.Rename(moveTo)

					if renameErr != nil {
						fmt.Printf("error renaming asset %s to %s: %v\n", moveFrom, moveTo, renameErr)
						os.Exit(-1)
					}
				} else {
					helpAssetCmd()
				}
			default:
				fmt.Printf("asset sub-command %s not understood\n", args[2])
				helpAssetCmd()
			}
		} else {
			helpAssetCmd()
			os.Exit(-1)
		}
	case "user":
		if len(args) >= 5 {
			switch strings.ToLower(args[2]) {
			case "add":
				username := args[3]
				emailAddress := args[4]

				fmt.Printf("Adding user (%s) with email address (%s)\n", username, emailAddress)

				grog := getModel()
				newUser := grog.NewUser(emailAddress, username)
				newUserErr := newUser.Save()
				if newUserErr != nil {
					fmt.Printf("error adding new user: %v\n", newUserErr)
				} else {
					fmt.Printf("user added\n")
				}

			default:
				fmt.Printf("user sub-command %s not understood\n", args[2])
				helpUserCmd()
			}
		} else {
			helpUserCmd()
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

		db := manageddb.NewManagedDB(dbFilename, "sqlite3", migrations.DatabaseMigrations, true)
		grog = model.NewModel(db)
	}

	return grog
}

func help() {
	fmt.Println("Usage:")
	helpAssetCmd()
	helpUserCmd()
}

func helpAssetCmd() {
	fmt.Printf("\tgrogcmd asset add [-ext] <file|directory>\n\t              mv <from> <to>\n")
}

func helpUserCmd() {
	fmt.Printf("\tgrogcmd user add \"name\" \"emailAddress\"\n")
}
