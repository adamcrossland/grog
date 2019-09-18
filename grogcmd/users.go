package main

import (
	"fmt"
	"os"
)

func listUsers() {
	grog := getModel()

	users, getUsersErr := grog.AllUsers()
	if getUsersErr != nil {
		fmt.Printf("error loading users from database: %v\n", getUsersErr)
		os.Exit(-1)
	}

	columnData := make([][]string, len(users))

	for i := 0; i < len(users); i++ {
		columnData[i] = make([]string, 4)
		columnData[i][0] = fmt.Sprintf("%d", users[i].ID)
		columnData[i][1] = users[i].Name
		columnData[i][2] = users[i].Email
		columnData[i][3] = fmt.Sprintf("%s %d %d", users[i].Added.Val().Month(),
			users[i].Added.Val().Day(), users[i].Added.Val().Year())
	}

	tabularOutput(columnData)
}
