package app

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func (s *server) routes(staticDir string) {

	registerStaticWeb(s.router, staticDir)

	s.router.Post("/upload", s.uploadFileHandler())
	s.router.Handle("/signup", s.signupHandler())
	s.router.HandleFunc("/about", s.handleAbout())
	// s.router.HandleFunc("/", s.handleIndex())
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (s *server) uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		fileType := r.PostFormValue("type")
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)
		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}

		fileName := randToken(12)
		// fileEndings, err := mime.ExtensionsByType(fileType)
		fileEndings := []string{".jpg"}

		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(s.receiveDir, fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func (s *server) signupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// boostrap is a template name defined in layout/boostrap.html
			v := NewView("bootstrap", "views/signup.html")
			v.Render(w, s.educatorNames)
		case "POST":
			if err := r.ParseForm(); err != nil {
				renderError(w, "CANNOT_PARSE_FORM", http.StatusBadRequest)
				return
			}
			fmt.Println("path", r.URL.Path)
			fmt.Println(r.Form) // print information on server side.

			educatorNames := r.Form["educators"]

			fmt.Printf("educator Names %s\n", educatorNames)

			v := NewView("bootstrap", "views/uploadFile.html")

			v.Render(w, educatorNames)
		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *server) handleAbout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Lepus version:%s", s.version)
	})
}

// func (s *server) handleIndex() http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
// 	})
// }
