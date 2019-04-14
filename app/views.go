package app

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type tempMap map[string]*template.Template

//initTemplates initialize  new templates by parsing go html template files
func (s *lepus) initTemplates(viewPath string) error {
	layoutFilePath := filepath.Join(viewPath, "layout/*.html")
	viewFilePath := filepath.Join(viewPath, "*.html")

	viewFiles, err := filepath.Glob(viewFilePath)
	if err != nil {
		logrus.WithError(err).Fatalf("filepath.Glob failed:%s", viewFilePath)
	}

	//get all layout files
	layoutFiles, err := filepath.Glob(layoutFilePath)
	if err != nil {
		logrus.WithError(err).Fatalf("filepath.Glob failed:%s", viewFilePath)
	}

	if layoutFiles == nil {
		logrus.Fatalf("Glob search did not find any layout file:%s", layoutFilePath)
	}

	s.tempMap = map[string]*template.Template{}

	for _, viewFile := range viewFiles {
		//verify template exiits
		_, err := os.Stat(viewFile)
		if os.IsNotExist(err) {
			logrus.Fatalf("html template file dose not exist : %s", viewFile)

		} else if err != nil {
			logrus.Fatalf("os error while check template file i %s:\n%v", viewFile, err)
		}

		// file exists
		files := append([]string{viewFile}, layoutFiles...)

		t, err := template.ParseFiles(files...)

		if t == nil {
			logrus.Fatalf("Did not find any template file matches:%s", files)
		}
		if err != nil {
			logrus.Fatalf("parse template failed:%s", files)
		}

		_, filename := filepath.Split(viewFile)
		s.tempMap[strings.TrimSuffix(filename, filepath.Ext(filename))] = t
	}
	logrus.Info("Template init completed.")
	return nil
}

//Render renders the template with input data and write result to w
func (s *lepus) Render(w http.ResponseWriter, urlPath string, data interface{}) error {
	t := s.tempMap[urlPath]
	if t == nil {
		logrus.Fatalf("cannot find template for url %s", urlPath)
	}
	return t.ExecuteTemplate(w, "bootstrap", data)
	// return nil
}
