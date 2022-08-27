package migrations

import (
	"database/sql"

	"github.com/adamcrossland/grog/manageddb"
)

// DatabaseMigrations contains all of the database migration changes.
var DatabaseMigrations map[int]manageddb.DBMigration

func init() {
	DatabaseMigrations = map[int]manageddb.DBMigration{
		1: {Up: migration1up, Down: migration1down},
		2: {Up: migration2up, Down: migration2down},
		3: {Up: migration3up, Down: migration3down},
	}
}

func migration1up(db *sql.DB) error {
	_, err := db.Exec("create table db_metadata (migration integer); insert into db_metadata (migration) values (0)")

	return err
}

func migration1down(db *sql.DB) error {
	_, err := db.Exec("drop table db_metadata")

	return err
}

func migration2up(db *sql.DB) error {
	var err error

	_, err = db.Exec(`CREATE TABLE users (
		ID	INTEGER NOT NULL,
		Email   TEXT NOT NULL,
		Name	TEXT NOT NULL,
		Added	NUMERIC NOT NULL,
		PRIMARY KEY('ID')
	)`)

	_, err = db.Exec(`create table content (id integer primary key,
		title text,
		summary text,
		body text,
		slug text,
		template text,
		parent integer,
		author integer,
		added numeric,
		modified numeric)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`create index slugindex on content (slug)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`create table assets (name text primary key,
		mimeType text,
		content blob,
		serve_external integer,
		rendered integer,
		added numeric,
		modified numeric)`)
	if err != nil {
		return err
	}

	return err
}

func migration2down(db *sql.DB) error {
	var err error

	_, err = db.Exec("drop table content")
	if err != nil {
		return err
	}

	_, err = db.Exec(`drop index slugindex`)
	if err != nil {
		return err
	}

	_, err = db.Exec("drop table assets")
	if err != nil {
		return err
	}

	_, err = db.Exec(`DROP TABLE users`)

	return err
}

func migration3up(db *sql.DB) error {
	var err error

	_, err = db.Exec(`CREATE TABLE "queries" (
		"id"	INTEGER PRIMARY KEY AUTOINCREMENT,
		"name"	TEXT NOT NULL,
		"query"	TEXT NOT NULL,
		"added"	NUMERIC,
		"modified"	NUMERIC
	);`)

	return err
}

func migration3down(db *sql.DB) error {
	var err error

	_, err = db.Exec("drop table queries")

	return err
}
