package app

import (
	"html/template"
	"net/http"
	"path/filepath"
)

const layoutFilePath = "views/layout/*.html"

//View may be exposed
type View struct {
	Template *template.Template
	Layout   string
}

func layoutFiles() []string {
	files, err := filepath.Glob(layoutFilePath)
	if err != nil {
		panic(err)
	}
	return files
}

//NewView creates new templates by parsing go html template files
func NewView(layout string, files ...string) *View {
	files = append(files, layoutFiles()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

//Render renders the template with input data and write result to w
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	// v.Template.ExecuteTemplate(os.Stderr, v.Layout, data)
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}
