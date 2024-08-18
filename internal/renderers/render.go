package renderers

import (
	"bytes"
	"fmt"
	"github.com/justinas/nosurf"
	"hotel_management_system/internal/config"
	"hotel_management_system/internal/models"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}

var customFunc = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	// PopString session mai data insert karke nikal deta hai
	// Idhar hum initialisation ke liye use kar rahe
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	} else {
		td.IsAuthenticated = 0
	}
	return td
}

func RenderTemplate(w http.ResponseWriter, tmpl string) {
	ts, err := template.ParseGlob("./templates/" + tmpl)
	if err != nil {
		fmt.Println("Error in parsing template.", err)
		return
	}
	err = ts.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
}

func RenderTemplateWithLayout(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) {
	var tempCache map[string]*template.Template
	var err error
	if app.UseCache {
		tempCache = app.TemplateCache
	} else {
		tempCache, err = CreateTemplateCache()
		if err != nil {
			fmt.Println("Error creating template cache:", err)
		}
	}
	//fmt.Println(tempCache)
	tmplt, ok := tempCache[tmpl]
	if !ok {
		fmt.Println("page not in temp cache:", err)
	}
	buf := new(bytes.Buffer)
	//fmt.Println(tmplt)

	templateData = AddDefaultData(templateData, r)

	err = tmplt.Execute(buf, templateData)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing from buf:", err)
		return
	}

}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := make(map[string]*template.Template)

	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		fmt.Println("error in finding pages:")
		return myCache, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		// creating name helps us search for template with just filename without the whole path
		ts, err := template.New(name).Funcs(customFunc).ParseFiles(page) //ts== template set
		if err != nil {
			fmt.Println("Error in parsing template.", err)
			return myCache, err
		}

		layouts, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			fmt.Println("Error in finding layouts:", err)
			return myCache, err
		}
		//fmt.Println("layouts:", layouts)
		if len(layouts) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				fmt.Println("Error in parsing layouts:", err)
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
