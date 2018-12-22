package model

import (
	"os"
	"testing"

	"github.com/adamcrossland/grog/manageddb"
	"github.com/adamcrossland/grog/migrations"
)

var testdbname = "./model_test.db"

func dbSetup() *manageddb.ManagedDB {
	os.Remove(testdbname)
	return manageddb.NewManagedDB(testdbname, "sqlite3", migrations.DatabaseMigrations, true)
}

func dbTeardown() {
	os.Remove(testdbname)
}
func TestAddPost(t *testing.T) {
	model := NewModel(dbSetup())

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

	dbTeardown()
}

func TestAddAsset(t *testing.T) {
	model := NewModel(dbSetup())

	testText := "This is a text file."
	newAsset := model.NewAsset("test.txt", "text/plain")
	newAsset.Write([]byte(testText))
	saveErr := newAsset.Save()

	if saveErr != nil {
		t.Fatalf("Saving new Asset resulted in database error: %v", saveErr)
	}

	savedAsset, loadErr := model.GetAsset("test.txt")

	if loadErr != nil {
		t.Fatalf("Getting just-saved Asset resulted in database error: %v", loadErr)
	}
	if savedAsset.Name != newAsset.Name {
		t.Fatalf("savedAsset had different name (%s) than newAsset (%s)", savedAsset.Name, newAsset.Name)
	}
	if savedAsset.Size() != len(testText) {
		t.Fatalf("savedAsset Content size was wrong, expected %d got %d", len(testText), savedAsset.Size())
	}
	savedBlob := string(savedAsset.Content[:])
	if savedBlob != testText {
		t.Fatalf("savedAsset Content did not match stored blob. Expected '%s', got '%s'", testText, savedBlob)
	}
	if savedAsset.MimeType != newAsset.MimeType {
		t.Fatal("savedAsset and newAsset had different MimeType values")
	}
	if savedAsset.Added.Unix() != newAsset.Added.Unix() {
		t.Fatalf("savedAsset and newAsset had different Added values (%d vs %d)", savedAsset.Added.Unix(),
			newAsset.Added.Unix())
	}
	if savedAsset.Modified.Unix() != newAsset.Modified.Unix() {
		t.Fatalf("savedAsset and newAsset had different Modified values (%d) vs (%d)", savedAsset.Modified.Unix(),
			newAsset.Modified.Unix())
	}

	dbTeardown()
}
