package model

import (
	"os"
	"testing"

	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/migrations"
)

var model *GrogModel
var testdbname = "./post_test.db"

func TestMain(m *testing.M) {
	os.Remove(testdbname)
	db := manageddb.NewManagedDB(testdbname, "sqlite3", migrations.DatabaseMigrations)

	model = NewModel(db)

	runResult := m.Run()

	os.Remove(testdbname)

	os.Exit(runResult)
}

func TestAddPost(t *testing.T) {
	newPost := model.NewPost("Test title", "Automated test post",
		"This post is created by automted testing and should never survive the testing process")
	saveErr := newPost.Save()

	if saveErr != nil {
		t.Fatalf("Saving new Post resulted in database error: %v", saveErr)
	}

	if !newPost.IndexSet() {
		t.Fatal("Id not set after post save.")
	}

	savedPost, loadErr := model.GetPost(newPost.ID)

	if loadErr != nil {
		t.Fatalf("Getting just-saved Post resulted in database error: %v", loadErr)
	}
	if !savedPost.IndexSet() {
		t.Fatal("savePost id was not set")
	}
	if savedPost.ID != newPost.ID {
		t.Fatalf("savePost had different id (%d) than newPost (%d)", savedPost.ID, newPost.ID)
	}
	if savedPost.Title != newPost.Title {
		t.Fatal("savedPost and newPost had different Title values")
	}
	if savedPost.Summary != newPost.Summary {
		t.Fatal("savedPost and newPost had different Summary values")
	}
	if savedPost.Body != newPost.Body {
		t.Fatal("savedPost and newPost had different Body values")
	}
	if savedPost.Added.Unix() != newPost.Added.Unix() {
		t.Fatalf("savedPost and newPost had different Added values (%d vs %d)", savedPost.Added.Unix(), newPost.Added.Unix())
	}
	if savedPost.Edited.Unix() != newPost.Edited.Unix() {
		t.Fatal("savedPost and newPost had different Edited values")
	}
}
