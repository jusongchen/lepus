package app

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

/*
	routing path:
	at home / form post -> /signup form post -> /selectphoto form post ->  where2 form post, branching to:
			1) home /
			2) /selectphoto
			3) /sponsor , post -> home /
*/
func (s *lepus) routes(staticDir string) {

	// A good base middleware stack
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}
	s.router.Use(NewStructuredLogger(logger))
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)

	registerStaticWeb(s.router, staticDir)

	s.router.Handle("/signup", s.signupHandler())
	s.router.Handle("/selectphoto", s.selectPhotoHandler())
	// s.router.Post("/upload", s.uploadHandler())
	s.router.Post("/where2", s.where2Handler())
	s.router.HandleFunc("/about", s.handleAbout())
	// s.router.HandleFunc("/", s.handleIndex())

}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	// validate file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(maxUploadSize)

	if err != nil {
		errMsg := fmt.Sprintf("上传的文件太大（已超过%d兆字节）", maxUploadSize/1024/1024)
		renderError(w, errMsg, http.StatusBadRequest)
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

}

// func (s *lepus) uploadHandler() http.HandlerFunc {
// 	return http.HandlerFunc()
// }

func (s *lepus) signupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s.Render(w, "signup", s.educatorNames)

		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) selectPhotoHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// case "GET":
		case "POST":

			if err := r.ParseForm(); err != nil {
				renderError(w, "CANNOT_PARSE_FORM", http.StatusBadRequest)
				return
			}
			sessionID := s.getSessionID(w, r)

			profile, err := getParticipantProfile(sessionID)
			if err != nil {
				return
			}

			log.Printf("Participant Profile at %v:%+v", r.URL, profile)

			data := struct {
				EducatorNames []string
				SessionID     string
			}{
				EducatorNames: profile.SelectedEducators,
				SessionID:     sessionID,
			}
			s.Render(w, "selectphoto", data)
		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) where2Handler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		// case "GET":
		case "POST":

			uploadFile(w, r)
			sessionID := s.getSessionID(w, r)

			s.Render(w, "where2", struct{ SessionID string }{SessionID: sessionID})
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
