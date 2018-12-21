package model

import "github.com/adamcrossland/grog/manageddb"

type GrogModel struct {
	db *manageddb.ManagedDB
}

// NewModel create a new GrogModel instance.
func NewModel(db *manageddb.ManagedDB) *GrogModel {
	newModel := new(GrogModel)
	newModel.db = db

	return newModel
}
