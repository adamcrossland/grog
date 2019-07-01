package model

import (
	"database/sql"
	"fmt"
	"time"
)

// User comprises all of the information about a user of Grog
type User struct {
	model *GrogModel
	ID    int64
	Email string
	Name  string
	Added NullTime
}

// NewUser creates a new instance of the User type
func (model *GrogModel) NewUser(email string, name string) *User {
	newUser := new(User)
	newUser.model = model
	newUser.ID = -1
	newUser.Email = email
	newUser.Name = name

	return newUser
}

// Save writes the User object to the database
func (user *User) Save() error {
	var saveError error

	if user.ID == -1 {
		// New, do insert
		if user.Added.IsNull() {
			user.Added.Set(time.Now())
		}

		insertResult, err := user.model.db.DB.Exec("insert into users (Email, Name, Added) values (?, ?, strftime('%s','now'))",
			user.Email, user.Name)
		if err == nil {
			user.ID, err = insertResult.LastInsertId()
		}

		saveError = err
	} else {
		// Exists, do update
		_, err := user.model.db.DB.Exec(`update users set Email = ?, Name = ? where Id = ?`,
			user.Email, user.Name, user.Added.Unix(), user.ID)
		saveError = err
	}

	return saveError
}

// GetUser retrieves the User object identified by the given id from the database
func (model *GrogModel) GetUser(id int64) (*User, error) {
	var foundUser *User
	var err error
	var email string
	var name string
	var added int64

	row := model.db.DB.QueryRow("select email, name, added from Users where id = ?", id)

	if row.Scan(&email, &name, &added) != sql.ErrNoRows {
		foundUser = model.NewUser(email, name)
		foundUser.ID = id
		foundUser.Added.Set(time.Unix(added, 0))
	} else {
		err = fmt.Errorf("No user with id %d", id)
	}

	return foundUser, err
}

// Delete removes the user with the given id from the database
func (model *GrogModel) DeleteUser(id int64) error {
	_, err := model.db.DB.Exec("delete from users where id=?", id)

	return err
}
