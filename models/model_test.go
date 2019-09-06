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

	newPost := model.NewContent("Test title", "Automated test post",
		"This post is created by automted testing and should never survive the testing process", "", "")
	saveErr := newPost.Save()

	if saveErr != nil {
		t.Fatalf("Saving new Post resulted in database error: %v", saveErr)
	}

	if !newPost.IndexSet() {
		t.Fatal("Id not set after post save.")
	}

	savedPost, loadErr := model.GetContent(newPost.ID)

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
		t.Fatalf("savedPost and newPost had different Title values (%s), (%s)", savedPost.Title, newPost.Title)
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
	if savedPost.Modified.Unix() != newPost.Modified.Unix() {
		t.Fatal("savedPost and newPost had different Edited values")
	}

	dbTeardown()
}
func TestAddAsset(t *testing.T) {
	model := NewModel(dbSetup())

	testText := "This is a text file."
	newAsset := model.NewAsset("test.txt", "text/plain")
	newAsset.Write([]byte(testText))
	newAsset.ServeExternal = true
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
	if !savedAsset.ServeExternal {
		t.Fatalf("savedAsset should have ServeExternal of true but does not.")
	}

	dbTeardown()
}
func TestPostSlugging(t *testing.T) {
	model := NewModel(dbSetup())

	testPost1 := model.NewContent("This is a test post", "", "TEST", "", "")

	if testPost1.Slug != "this-is-a-test-post" {
		t.Fatalf("testPost1 has incorrect slug %s", testPost1.Slug)
	}

	testPost2 := model.NewContent("This   is a test   post", "", "TEST", "", "")

	if testPost2.Slug != "this-is-a-test-post" {
		t.Fatalf("testPost2 has incorrect slug %s", testPost2.Slug)
	}

	testPost3 := model.NewContent("This is @a test' post!", "", "TEST", "", "")

	if testPost3.Slug != "this-is-a-test-post" {
		t.Fatalf("testPost3 has incorrect slug %s", testPost3.Slug)
	}

	dbTeardown()
}
func TestAddingUser(t *testing.T) {
	model := NewModel(dbSetup())

	testUser := model.NewUser("testuser@test.com", "Test User")
	saveErr := testUser.Save()

	if saveErr != nil {
		t.Fatalf("saving of testUser failed with error: %v", saveErr)
	}

	if testUser.ID == -1 {
		t.Fatal("testUser did not receive a valid ID after being saved")
	}

	foundUser, foundErr := model.GetUser(testUser.ID)

	if foundErr != nil {
		t.Fatalf("GetUser failed with error: %v", foundErr)
	}

	if foundUser.ID != testUser.ID {
		t.Fatalf("testUser.ID (%d) and foundUser.ID (%d) do not match", testUser.ID, foundUser.ID)
	}

	if foundUser.Email != testUser.Email {
		t.Fatalf("testUser.Email (%s) and foundUser.Email (%s) do not match", testUser.Email, foundUser.Email)
	}

	if foundUser.Name != testUser.Name {
		t.Fatalf("testUser.Name (%s) and foundUser.Name (%s) do not match", testUser.Name, foundUser.Name)
	}

	if foundUser.Added.Unix() != testUser.Added.Unix() {
		t.Fatalf("testUser.Added (%d) and foundUser.Added (%d) do not match", testUser.Added.Unix(), foundUser.Added.Unix())
	}

	dbTeardown()
}

func TestContentChildren(t *testing.T) {
	model := NewModel(dbSetup())

	newPost := model.NewContent("Test Parent", "Automated test post with children",
		"This post is created by automted testing to test the ability to load child content", "", "")
	saveErr := newPost.Save()

	if saveErr != nil {
		t.Fatalf("Saving new Content resulted in database error: %v", saveErr)
	}

	if !newPost.IndexSet() {
		t.Fatal("Id not set after Content save.")
	}

	childPost1 := model.NewContent("Child 1", "Child Content 1",
		"This content should be the first child of Test Parent", "", "")
	childPost1.Parent = newPost.ID
	saveErr = childPost1.Save()
	if saveErr != nil {
		t.Fatalf("Saving Child 1 Content resulted in database error: %v", saveErr)
	}
	if !childPost1.IndexSet() {
		t.Fatal("Id not set after Child 1 Content save.")
	}

	childPost2 := model.NewContent("Child 2", "Child Content 2",
		"This content should be the second child of Test Parent", "", "")
	childPost2.Parent = newPost.ID
	saveErr = childPost2.Save()
	if saveErr != nil {
		t.Fatalf("Saving Child 2 Content resulted in database error: %v", saveErr)
	}
	if !childPost2.IndexSet() {
		t.Fatal("Id not set after Child 2 Content save.")
	}

	childPost3 := model.NewContent("Child 3", "Child Content 3",
		"This content should be the third child of Test Parent", "", "")
	childPost3.Parent = newPost.ID
	saveErr = childPost3.Save()
	if saveErr != nil {
		t.Fatalf("Saving Child 3 Content resulted in database error: %v", saveErr)
	}
	if !childPost3.IndexSet() {
		t.Fatal("Id not set after Child 3 Content save.")
	}

	if len(newPost.Children) != 0 {
		t.Fatalf("newPost Children was %d instead of 0", len(newPost.Children))
	}

	newPost.IncludeChildren()

	if len(newPost.Children) != 3 {
		t.Fatalf("newPost Children was %d instead of 3", len(newPost.Children))
	}

	//dbTeardown()
}
