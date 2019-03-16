package model

import (
	"database/sql"
	"fmt"
	"time"
)

// Asset is a file this is or will be stored in the database
type Asset struct {
	model         *GrogModel
	Name          string
	MimeType      string
	Content       []byte
	ServeExternal bool
	Rendered      bool
	Added         NullTime
	Modified      NullTime
}

// NewAsset creates a new Asset object
func (model *GrogModel) NewAsset(name string, mimeType string) *Asset {

	newAsset := new(Asset)
	newAsset.Name = name
	newAsset.MimeType = mimeType
	newAsset.model = model

	return newAsset
}

// GetAsset loads the Asset with the given name from the database
func (model *GrogModel) GetAsset(name string) (*Asset, error) {
	var foundAsset *Asset
	var mimeType string
	var content = make([]byte, 0)
	var serveExternal int64
	var rendered int64
	var added int64
	var modified int64
	var err error

	row := model.db.DB.QueryRow(`select mimeType, content, serve_external, rendered,
		added, modified from Assets where name = ?`, name)
	if row.Scan(&mimeType, &content, &serveExternal, &rendered, &added, &modified) != sql.ErrNoRows {
		foundAsset = model.NewAsset(name, mimeType)
		foundAsset.Content = content
		if serveExternal == 1 {
			foundAsset.ServeExternal = true
		} else {
			foundAsset.ServeExternal = false
		}

		if rendered == 1 {
			foundAsset.Rendered = true
		} else {
			foundAsset.Rendered = false
		}

		foundAsset.Added.Set(time.Unix(added, 0))
		foundAsset.Modified.Set(time.Unix(modified, 0))
	} else {
		err = fmt.Errorf("No asset with name %s", name)
	}

	return foundAsset, err
}

// Exists checks to see if an asset by this name is already stored in the database
func (asset Asset) Exists() bool {
	doesExist := false

	row := asset.model.db.DB.QueryRow("select count(1) from Assets where name = ?", asset.Name)
	var count int
	row.Scan(&count)
	if count > 0 {
		doesExist = true
	}

	return doesExist
}

// Save stores the Asset in the database
func (asset *Asset) Save() error {
	var saveError error

	var serveExternalVal int64
	if asset.ServeExternal {
		serveExternalVal = 1
	}

	var renderedVal int64
	if asset.Rendered {
		renderedVal = 1
	}

	if !asset.Exists() {
		// New, do insert
		if asset.Added.IsNull() {
			asset.Added.Set(time.Now())
		}

		if asset.Modified.IsNull() {
			asset.Modified.Set(asset.Added.Time)
		}

		_, err := asset.model.db.DB.Exec(`insert into assets (name, mimeType, content, serve_external,
			rendered, added, modified) values (?, ?, ?, ?, ?, ?, ?)`,
			asset.Name, asset.MimeType, asset.Content, serveExternalVal, renderedVal,
			asset.Added.Unix(), asset.Modified.Unix())

		saveError = err
	} else {
		// Exists, do update

		asset.Modified.Set(time.Now())

		_, err := asset.model.db.DB.Exec(`update assets set mimeType = ?, content = ?, serve_external = ?,
				rendered = ?, modified = ? where name = ?`, asset.MimeType, asset.Content, serveExternalVal,
			renderedVal, asset.Modified.Unix(), asset.Name)
		saveError = err
	}

	return saveError
}

func (asset *Asset) Write(p []byte) (n int, err error) {
	asset.Content = make([]byte, len(p))
	copy(asset.Content, p)

	return len(p), nil
}

func (asset Asset) Read(p []byte) (n int, err error) {
	copy(asset.Content, p)

	return len(asset.Content), nil
}

// Size returns the number of bytes stored in the Asset's Content.
func (asset Asset) Size() int {
	return len(asset.Content)
}
