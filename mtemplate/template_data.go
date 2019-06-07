package mtemplate

import (
	"net/http"

	model "github.com/adamcrossland/grog/models"
)

// TemplateData comprises all of the information required for processing template rendering
type TemplateData struct {
	data         map[string]interface{} // Hold all data passed in for rendering the template
	response     http.ResponseWriter
	request      *http.Request
	NamedQueries map[string]model.NamedQueryFunc
}

// NewTemplateData returns a new TemplateData object
func NewTemplateData(wr http.ResponseWriter, r *http.Request, namedQueries map[string]model.NamedQueryFunc,
	data interface{}) *TemplateData {
	newData := new(TemplateData)

	newData.response = wr
	newData.request = r
	newData.data = make(map[string]interface{})

	newData.data["model"] = data
	newData.NamedQueries = namedQueries

	return newData
}

// SetCookie sets a cookie in the response
func (data *TemplateData) SetCookie(cookie *http.Cookie) {
	http.SetCookie(data.response, cookie)
}

// GetCookie gets a cookie from the request
func (data *TemplateData) GetCookie(name string) (*http.Cookie, error) {
	return data.request.Cookie(name)
}
