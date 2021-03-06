package manageddb

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

// ManagedDB has all information pertaining to the database that is being managed.
// This should probably all be opaque to the user.
type ManagedDB struct {
	DB               *sql.DB
	dbLock           *sync.Mutex
	migrations       map[int]DBMigration
	currentMigration int
	silent           bool
}

// NewManagedDB creates and initializes a new ManagedDB with the given file path,
// datatbase driver and migrations to apply to the db.
func NewManagedDB(dbPath string, driver string, migrations map[int]DBMigration, silent bool) *ManagedDB {
	newDB := new(ManagedDB)
	var err error
	newDB.silent = silent

	newDB.DB, err = sql.Open(driver, dbPath)
	if err != nil {
		panic(fmt.Sprintf("err opening database: %v", err))
	}

	newDB.dbLock = new(sync.Mutex)

	// Figure out what the current migration is
	newDB.currentMigration = newDB.getCurrentMigration()
	if !silent {
		log.Printf("Current migration level: %d.", newDB.currentMigration)
	}

	newDB.migrations = migrations

	newDB.databaseMigrate(-1)
	return newDB
}

func (mdb ManagedDB) getCurrentMigration() int {
	var current int

	rows, err := mdb.DB.Query("select migration from db_metadata")
	if err == nil {
		defer rows.Close()
		rows.Next()
		if err = rows.Scan(&current); err != nil {
			log.Fatalf("Error getting current migration from db: %v", err)
			panic(err)
		}
	}

	return current
}

func (mdb ManagedDB) setCurrentMigration(level int) {
	_, err := mdb.DB.Exec("update db_metadata set migration = ?", level)
	if err != nil {
		log.Fatalf("unable to update database migration level: %v", err)
		panic(err)
	}
}

func (mdb ManagedDB) databaseMigrate(toMigration int) {
	// If desired migration level is -1, it means go to the latest
	// migration.
	if toMigration == -1 {
		toMigration = len(mdb.migrations)
	}

	var dbErr error

	if mdb.currentMigration > toMigration {
		// Migrating down.
		for mdb.currentMigration > toMigration {
			mdb.currentMigration--
			dbErr = mdb.migrations[mdb.currentMigration].Down(mdb.DB)
			if dbErr != nil {
				panic(fmt.Sprintf("db migration %d down failed: %v", mdb.currentMigration, dbErr))
			} else {
				mdb.setCurrentMigration(mdb.currentMigration)
			}
		}
	} else if mdb.currentMigration < toMigration {
		// Migrating up.
		for mdb.currentMigration < toMigration {
			mdb.currentMigration++
			dbErr = mdb.migrations[mdb.currentMigration].Up(mdb.DB)
			if dbErr != nil {
				panic(fmt.Sprintf("db migration %d up failed: %v", mdb.currentMigration, dbErr))
			} else {
				mdb.setCurrentMigration(mdb.currentMigration)
			}
		}

		if !mdb.silent {
			log.Printf("Migrated up to level %d.", mdb.currentMigration)
		}
	} else {
		if !mdb.silent {
			log.Printf("No migrations to perform.")
		}
	}
}

// WriteFunc is a function signature that should be implemented by any function
// that performs write operations to the database.
type WriteFunc func(db *sql.DB) error

// DoWrite executes the provided ManagedDBWriteFunc in a safely
// single-threaded way. All writes to the underlying DB should happen
// in this way.
func (mdb ManagedDB) DoWrite(writeFunc WriteFunc) error {
	mdb.dbLock.Lock()
	writeErr := writeFunc(mdb.DB)
	mdb.dbLock.Unlock()

	return writeErr
}

// DBMigrationFunction gives the signature of functions that can perform
// database migrations.
type DBMigrationFunction func(db *sql.DB) error

// DBMigration contains two functions, Up and Down that together perform and undo
// a set of changes to the database.
type DBMigration struct {
	Up   DBMigrationFunction
	Down DBMigrationFunction
}
