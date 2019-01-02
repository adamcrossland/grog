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
		"This post is created by automted testing and should never survive the testing process", "")
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

	testPost1 := model.NewPost("This is a test post", "", "TEST", "")

	if testPost1.Slug != "this-is-a-test-post" {
		t.Fatalf("testPost1 has incorrect slug %s", testPost1.Slug)
	}

	testPost2 := model.NewPost("This   is a test   post", "", "TEST", "")

	if testPost2.Slug != "this-is-a-test-post" {
		t.Fatalf("testPost2 has incorrect slug %s", testPost2.Slug)
	}

	testPost3 := model.NewPost("This is @a test' post!", "", "TEST", "")

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

func TestAddComment(t *testing.T) {
	model := NewModel(dbSetup())

	testUser := model.NewUser("testuser@test.com", "Test User")
	testUser.Save()

	newPost := model.NewPost("Test title", "Automated test post",
		"This post is created by automted testing and should never survive the testing process", "")
	newPost.Save()

	newComment, err := newPost.AddComment("This is a test comment.", *testUser)
	if err != nil {
		t.Fatalf("error adding comment: %v", err)
	}

	if newComment == nil {
		t.Fatal("newCOmment was nil")
	}

	if newComment.ID == -1 {
		t.Fatal("newComment was not assigned a valid ID")
	}

	dbTeardown()
}

func TestGetComments(t *testing.T) {
	model := NewModel(dbSetup())

	testUser := model.NewUser("testuser@test.com", "Test User")
	testUser.Save()

	newPost := model.NewPost("Test title", "Automated test post",
		"This post is created by automted testing and should never survive the testing process", "")
	newPost.Save()

	newPost.AddComment("This is a test comment.", *testUser)
	newPost.AddComment("This is a second test comment.", *testUser)
	newPost.AddComment("Testing is useful.", *testUser)

	postComments, err := newPost.LoadComments()
	if err != nil {
		t.Fatalf("error retrieving comments for post: %v", err)
	}
	if postComments == nil {
		t.Fatal("comments for post was nil")
	}
	if len(postComments) != 3 {
		t.Fatalf("expected 3 comments on test post, but got %d", len(postComments))
	}
	if postComments[0].Content != "This is a test comment." {
		t.Fatalf("incorrect content for comment 0: %s", postComments[0].Content)
	}
	if postComments[0].Author != testUser.ID {
		t.Fatalf("incorrect author for comment 0. expected %d, got %d", testUser.ID, postComments[0].Author)
	}
	if postComments[2].Content != "Testing is useful." {
		t.Fatalf("incorrect content for comment 2: %s", postComments[2].Content)
	}
	if newPost.Comments == nil {
		t.Fatalf("LoadComments did not save comments to Comments field")
	}
	if len(newPost.Comments) != 3 {
		t.Fatalf("expected 3 comments in newPost.Comments, but got %d", len(newPost.Comments))
	}
	if newPost.Comments[1].Content != "This is a second test comment." {
		t.Fatalf("incorrect content for newPost.Comments[1]")
	}
	if newPost.Comments[2].AuthorName != "Test User" {
		t.Fatalf("incorrect authorname. expected 'Test User' but got '%s", newPost.Comments[2].AuthorName)
	}
	if newPost.Comments[2].AuthorEmail != "testuser@test.com" {
		t.Fatalf("incorrect authoremail. expected 'testuser@test.com' but got '%s'", newPost.Comments[2].AuthorEmail)
	}

	dbTeardown()
}
