package app

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type tempMap map[string]*template.Template

var initViewFilenames = map[string]string{

	"selectphoto": "selectphoto.html",
	"signup":      "signup.html",
	"where2":      "where2.html",
}

//initTemplates initialize  new templates by parsing go html template files
func (s *lepus) initTemplates(viewPath string) error {
	layoutFilePath := filepath.Join(viewPath, "layout/*.html")

	viewFiles := map[string]string{}

	for k, v := range initViewFilenames {
		viewFiles[k] = filepath.Join(viewPath, v)
	}

	//get all layout files
	layoutFiles, err := filepath.Glob(layoutFilePath)
	if err != nil {
		panic(err)
	}

	if layoutFiles == nil {
		log.Fatalf("Glob search did not find any layout file:%s", layoutFilePath)
	}

	s.tempMap = map[string]*template.Template{}

	for k, v := range viewFiles {
		//verify template exiits
		_, err := os.Stat(v)
		if os.IsNotExist(err) {
			log.Fatalf("html template file dose not exist : %s", v)

		} else if err != nil {
			log.Fatalf("os error while check template file i %s:\n%v", v, err)
		}

		// file exists
		files := append([]string{v}, layoutFiles...)
		t, err := template.ParseFiles(files...)

		if t == nil {
			log.Fatalf("Did not find any template file matches:%s", layoutFilePath)
		}
		if err != nil {
			panic(err)
		}
		s.tempMap[k] = t
	}

	return nil
}

//Render renders the template with input data and write result to w
func (s *lepus) Render(w http.ResponseWriter, urlPath string, data interface{}) error {
	t := s.tempMap[urlPath]
	if t == nil {
		log.Fatalf("cannot find template for url %s", urlPath)
	}
	return t.ExecuteTemplate(w, "bootstrap", data)
	// return nil
}
