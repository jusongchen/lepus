package app

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type tempMap map[string]*template.Template

//initTemplates initialize  new templates by parsing go html template files
func (s *lepus) initTemplates(viewPath string) error {
	layoutFilePath := filepath.Join(viewPath, "layout/*.html")
	viewFilePath := filepath.Join(viewPath, "*.html")

	viewFiles, err := filepath.Glob(viewFilePath)
	if err != nil {
		log.WithError(err).Fatalf("filepath.Glob failed:%s", viewFilePath)
	}

	//get all layout files
	layoutFiles, err := filepath.Glob(layoutFilePath)
	if err != nil {
		log.WithError(err).Fatalf("filepath.Glob failed:%s", viewFilePath)
	}

	if layoutFiles == nil {
		log.Fatalf("Glob search did not find any layout file:%s", layoutFilePath)
	}

	s.tempMap = map[string]*template.Template{}

	for _, viewFile := range viewFiles {
		//verify template exiits
		_, err := os.Stat(viewFile)
		if os.IsNotExist(err) {
			log.Fatalf("html template file dose not exist : %s", viewFile)

		} else if err != nil {
			log.Fatalf("os error while check template file i %s:\n%v", viewFile, err)
		}

		// file exists
		files := append([]string{viewFile}, layoutFiles...)

		t, err := template.ParseFiles(files...)

		if t == nil {
			log.Fatalf("Did not find any template file matches:%s", files)
		}
		if err != nil {
			log.Fatalf("parse template failed:%s", files)
		}

		_, filename := filepath.Split(viewFile)
		s.tempMap[strings.TrimSuffix(filename, filepath.Ext(filename))] = t
	}
	log.Info("Template init completed.")
	return nil
}

//Render renders the template with input data and write result to w
func (s *lepus) Render(w http.ResponseWriter, urlPath string, data interface{}) error {
	t := s.tempMap[urlPath]
	if t == nil {
		log.Errorf("cannot find template for url %s", urlPath)
	}
	// t.ExecuteTemplate(os.Stderr, "bootstrap", data)
	return t.ExecuteTemplate(w, "bootstrap", data)
}
