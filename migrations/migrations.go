package migrations

import (
	"database/sql"

	"github.com/adamcrossland/grog/manageddb"
)

// DatabaseMigrations contains all of the database migration changes.
var DatabaseMigrations map[int]manageddb.DBMigration

func init() {
	DatabaseMigrations = map[int]manageddb.DBMigration{
		1: manageddb.DBMigration{Up: migration1up, Down: migration1down},
		2: manageddb.DBMigration{Up: migration2up, Down: migration2down},
		3: manageddb.DBMigration{Up: migration3up, Down: migration3down},
		4: manageddb.DBMigration{Up: migration4up, Down: migration4down},
		5: manageddb.DBMigration{Up: migration5up, Down: migration5down},
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
	_, err := db.Exec(`create table posts (id integer primary key,
		title text,
		summary text,
		body text,
		added integer,
		edited integer)`)

	return err
}

func migration2down(db *sql.DB) error {
	_, err := db.Exec("drop table posts")

	return err
}

func migration3up(db *sql.DB) error {
	_, err := db.Exec(`create table assets (name text primary key,
		mimeType text,
		content blob,
		serve_external integer,
		added integer,
		modified integer)`)

	return err
}

func migration3down(db *sql.DB) error {
	_, err := db.Exec("drop table assets")

	return err
}

func migration4up(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE "posts" ADD COLUMN slug TEXT`)

	return err
}

func migration4down(db *sql.DB) error {

	return nil
}

func migration5up(db *sql.DB) error {
	_, err := db.Exec(`CREATE INDEX slugindex on posts (slug)`)

	return err
}

func migration5down(db *sql.DB) error {
	_, err := db.Exec(`DROP INDEX slugindex on posts`)

	return err
}
