package app

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
)

func newSessionID(name string, gradYear string, educators []string) string {

	signupData := struct {
		Name              string
		GradYear          string
		SelectedEducators []string
	}{
		Name:              name,
		GradYear:          gradYear,
		SelectedEducators: educators,
	}

	b, err := json.Marshal(signupData)
	if err != nil {
		msg := fmt.Sprintf("json Marshall %v failed:%v", signupData, err)
		log.Fatalf(msg)
	}

	return string(b)

}

func (s *lepus) routes(staticDir string) {

	// A good base middleware stack
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	registerStaticWeb(s.router, staticDir)

	s.router.Post("/upload", s.uploadHandler())
	s.router.Handle("/signup", s.signupHandler())
	s.router.Handle("/selectphoto", s.selectPhotoHandler())
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

func (s *lepus) uploadHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, fmt.Sprintf("上传的文件太大（已超过%d兆字节）", maxUploadSize), http.StatusBadRequest)
			return
		}

		fmt.Printf("upload photo form value:%s\n", r.Form)

		// parse and validate file and post parameters
		fileType := r.PostFormValue("type")
		file, fileHeader, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, fmt.Sprintf("内部错误: r.FromFile get error:%s", err), http.StatusBadRequest)
			return
		}
		fmt.Printf("fileHeader: Filename %v Size %v", fileHeader.Filename, fileHeader.Size)

		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "内部错误，无法读取上传文件", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)
		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
			break
		default:
			renderError(w, "不认识的文件格式", http.StatusBadRequest)
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

		// w.Write([]byte("SUCCESS"))

		// sessionID := string(r.Form["sessionID"][0])
		// log.Printf("sessionID in upload:%+v", sessionID)
		// v := NewView("bootstrap", "views/uploadnext.html")
		// v.Render(w, sessionID)
		TODO here : Jusong Pass session to context
		
		http.Redirect(w, r, "/uploadnext", http.StatusSeeOther)
	})
}

func (s *lepus) signupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// boostrap is a template name defined in layout/boostrap.html
			v := NewView("bootstrap", "views/signup.html")
			v.Render(w, s.educatorNames)

		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) selectPhotoHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// case "GET":
		// boostrap is a template name defined in layout/boostrap.html
		// v := NewView("bootstrap", "views/signup.html")
		// v.Render(w, s.educatorNames)
		case "POST":

			if err := r.ParseForm(); err != nil {
				renderError(w, "CANNOT_PARSE_FORM", http.StatusBadRequest)
				return
			}

			fmt.Println("value get from", r.URL, r.Form) // print information on server side.
			sessionID := newSessionID(r.Form["name"][0], r.Form["gradYear"][0], r.Form["educators"])
			log.Printf("sessionID:%+v", sessionID)
			v := NewView("bootstrap", "views/selectphoto.html")
			data := struct {
				EducatorNames []string
				SessionID     string
			}{
				EducatorNames: r.Form["educators"],
				SessionID:     sessionID,
			}
			v.Render(w, data)
		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) handleAbout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Lepus version:%s", s.version)
	})
}

// func (s *LepusServer) handleIndex() http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 	})
// }
