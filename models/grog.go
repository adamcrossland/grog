package model

import "github.com/adamcrossland/grog/manageddb"

// GrogModel has all of the data and methods for interacting with the database that
// backs Grog.
type GrogModel struct {
	db *manageddb.ManagedDB
}

// NewModel create a new GrogModel instance.
func NewModel(db *manageddb.ManagedDB) *GrogModel {
	newModel := new(GrogModel)
	newModel.db = db

	return newModel
}
