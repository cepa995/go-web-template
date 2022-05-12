package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"reflect"
	"time"

	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/models"
	"github.com/justinas/nosurf"
)

//template.FuncMap is map of custom functions that we can use in a particular TEMPLATE (usually functions that are not built in the templating language)
var functions = template.FuncMap{
	"humanDate":   HumanDate,
	"isString":    IsString,
	"isInt":       IsInt,
	"isAvailable": IsAvailable,
}
var app *config.AppConfig
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// IsString returns true if argument is of type string
func IsString(i interface{}) bool {
	_, ok := i.(string)
	return ok
}

// IsInt returns true if argument is of type int
func IsInt(i interface{}) bool {
	_, ok := i.(int)
	return ok
}

// IsAvailable checks whether or not struct field is available
func IsAvailable(name string, data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	return v.FieldByName(name).IsValid()
}

// AddDefaultData creates default models.TemplateData which should be accessable to each template when rendered.
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CSRFToken = nosurf.Token(r)
	// If user signed in, auth token is automatically generated and stored in DB.
	// Here, we check if it exists in session, if it does we say that user is authenticated
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	if app.Session.Exists(r.Context(), "access_level") {
		access_level := app.Session.Get(r.Context(), "access_level").(int64)
		td.AccessLevel = access_level
	}
	return td
}

// Template renders template using html/template package.
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("cannot get template from cache")
	}

	td = AddDefaultData(td, r)

	buf := new(bytes.Buffer)
	err := t.Execute(buf, td)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}

// CreateTemplateCache creates a Template Cache map[string]*tempalte.Template{} which stores all application templates in memory
// and makes them easier to load; loading from internal memory is faster then loading from disk each time.
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Step 1. Select everything inside ./templates/ directory that starts with *.page.gohtml
	pages, err := filepath.Glob(fmt.Sprintf("%s/*page.gohtml", pathToTemplates))
	if err != nil {
		return nil, err
	}

	// Step 2. Parse each template page and store it in the map[string]*template.Template
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Step 2.1. Select everything inside ./templates/ directory that starts with *.layout.gohtml
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return nil, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
