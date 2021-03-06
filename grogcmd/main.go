package main

import (
	"fmt"
	"os"
	"strconv"
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
		if len(args) >= 3 {
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
						helpAssetCmd(false)
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
					helpAssetCmd(false)
				}
			case "set":
				props := make([]boolProperty, 0, 2)
				assetName := ""

				for _, paramVal := range args[3:] {
					if paramVal[0] == '-' || paramVal[0] == '+' {
						switch strings.ToLower(paramVal) {
						case "-ext":
							props = append(props, boolProperty{Name: "external", Value: false})
						case "+ext":
							props = append(props, boolProperty{Name: "external", Value: true})
						case "-render":
							props = append(props, boolProperty{Name: "render", Value: false})
						case "+render":
							props = append(props, boolProperty{Name: "render", Value: true})
						default:
							fmt.Printf("flag %s not understood\n", paramVal)
							helpAssetCmd(false)
							os.Exit(-1)
						}
					} else {
						// If not a flag, must be the filename to set
						assetName = paramVal
					}
				}

				if len(assetName) == 0 {
					fmt.Printf("must provide name of asset to set values on\n")
					helpAssetCmd(false)
					os.Exit(-1)
				}

				setAssetProps(assetName, props)
			case "ls":
				listAssets()
			case "update":
				if len(args) < 4 {
					fmt.Printf("asset update: too few parameters\n")
					helpAssetCmd(false)
					os.Exit(-1)
				}

				assetName := args[3]
				source := os.Stdin

				if len(args) >= 5 {
					// Get filename for updated data from command line
					var fileErr error
					source, fileErr = os.Open(args[4])
					if fileErr != nil {
						fmt.Printf("error opening file %s: %v\n", assetName, fileErr)
						os.Exit(-1)
					}
				}

				updateAsset(assetName, source)
			default:
				fmt.Printf("asset sub-command %s not understood\n", args[2])
				helpAssetCmd(false)
			}
		} else {
			helpAssetCmd(false)
			os.Exit(-1)
		}
	case "user":
		if len(args) >= 3 {
			switch strings.ToLower(args[2]) {
			case "add":
				if len(args) == 5 {
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
				} else {
					helpUserCmd(false)
					os.Exit(-1)
				}
			case "rm":
				if len(args) == 4 {
					userID := args[3]
					userIDInt, convErr := strconv.ParseInt(userID, 10, 64)
					if convErr != nil {
						fmt.Printf("userid parameter must be convertible to an integer\n")
						helpUserCmd(false)
						os.Exit(-1)
					}

					grog := getModel()
					delErr := grog.DeleteUser(userIDInt)
					if delErr != nil {
						fmt.Printf("error deleting user %d: %v\n", userIDInt, delErr)
						os.Exit(-1)
					}

					fmt.Printf("user %d deleted\n", userIDInt)
				} else {
					helpUserCmd(false)
					os.Exit(-1)
				}
			case "ls":
				listUsers()
			default:
				fmt.Printf("user sub-command %s not understood\n", args[2])
				helpUserCmd(false)
			}
		} else {
			helpUserCmd(false)
			os.Exit(-1)
		}
	case "content":
		switch strings.ToLower(args[2]) {
		case "ls":
			longList := false
			if len(args) > 3 {
				for _, lsOpt := range args[3:] {
					if lsOpt == "-l" {
						longList = true
					}
				}
			}

			listContent(longList)
		case "add":
			addContent(os.Stdin)
		case "update":
			if len(args) < 4 {
				fmt.Printf("content update: too few parameters\n")
				helpContentCmd(false)
				os.Exit(-1)
			}

			contentID := args[3]
			source := os.Stdin

			if len(args) >= 5 {
				// Get filename for updated data from command line
				var fileErr error
				source, fileErr = os.Open(args[4])
				if fileErr != nil {
					fmt.Printf("error opening file %s: %v\n", args[4], fileErr)
					os.Exit(-1)
				}
			}

			contentIDInt, convErr := strconv.ParseInt(contentID, 10, 64)
			if convErr != nil {
				fmt.Printf("contentid must be convertible to an integer\n")
				helpContentCmd(false)
				os.Exit(-1)
			}

			updateContentBody(contentIDInt, source)
		case "rm":
			if len(args) < 3 {
				fmt.Printf("content rm: too few parameters\n")
				helpContentCmd(false)
				os.Exit(-1)
			} 

			contentID := args[3]
			contentIDInt, convErr := strconv.ParseInt(contentID, 10, 64)
			if convErr != nil {
				fmt.Printf("contentid must be convertible to an integer\n")
				helpContentCmd(false)
				os.Exit(-1)
			}

			deleteContent(contentIDInt)
		default:
			helpContentCmd(false)
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
	helpAssetCmd(true)
	helpContentCmd(true)
	helpUserCmd(true)
}

func helpAssetCmd(usageShown bool) {
	if !usageShown {
		fmt.Println("Usage:")
	}
	fmt.Printf("\tgrogcmd asset add [-ext] <file|directory>\n")
	fmt.Printf("\t              mv <from> <to>\n")
	fmt.Printf("\t              set [+-ext] [+-render] <file|directory>\n")
	fmt.Printf("\t              ls\n")
	fmt.Printf("\t				update <assetname> [filename]\n")
	fmt.Println()
}

func helpUserCmd(usageShown bool) {
	if !usageShown {
		fmt.Println("Usage:")
	}
	fmt.Printf("\tgrogcmd user add \"name\" \"emailAddress\"\n")
	fmt.Printf("\t             rm id\n")
	fmt.Printf("\t             ls\n")
	fmt.Println()
}

func helpContentCmd(usageShown bool) {
	if !usageShown {
		fmt.Println("Usage:")
	}
	fmt.Printf("\tgrogcmd content ls -l\n")
	fmt.Printf("\t                add\n")
	fmt.Printf("\t                set contentid [template=templatename] [parent=parentid] [author=authorid]\n")
	fmt.Printf("\t                update contentid [filename]\n")
	fmt.Printf("\t                rm contentid\n")
	fmt.Println()
}
