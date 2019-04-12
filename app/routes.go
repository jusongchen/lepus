package app

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/h2non/filetype"
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

func uploadFile(w http.ResponseWriter, r *http.Request) error {

	var err error
	// validate file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err = r.ParseMultipartForm(maxUploadSize)

	if err != nil {
		errMsg := fmt.Sprintf("上传的文件太大（已超过%d兆字节）", maxUploadSize/1024/1024)
		renderError(w, errMsg, http.StatusBadRequest)
		return err
	}

	fmt.Printf("upload photo form value:%s\n", r.Form)

	// parse and validate file and post parameters
	// fileType := r.PostFormValue("type")

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		renderError(w, fmt.Sprintf("内部错误: r.FromFile get error:%s", err), http.StatusBadRequest)
		return err
	}
	fmt.Printf("fileHeader: Filename %v Size %v", fileHeader.Filename, fileHeader.Size)

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "内部错误，无法读取上传文件", http.StatusBadRequest)
		return err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	// filetyp := http.DetectContentType(fileBytes)
	// fileTypeExtensions := map[string]string{
	// 	"image/jpeg":      ".jpeg",
	// 	"image/jpg":       ".jpg",
	// 	"image/gif":       ".gif",
	// 	"image/png":       ".png",
	// 	"video/x-flv":     ".flv",
	// 	"video/mp4":       ".mp4",
	// 	"video/3gpp":      ".3gpp",
	// 	"video/quicktime": ".mov",
	// 	"video/x-msvideo": ".avi",
	// 	"video/x-ms-wmv":  ".wmv",
	// }

	fileExtensions := []string{".jpeg", ".jpg", ".gif", ".png", ".flv", ".mp4", ".3gpp", ".mov", ".avi", ".wmv"}

	if !filetype.IsImage(fileBytes) && !filetype.IsVideo(fileBytes) && !!filetype.IsAudio(fileBytes) {
		err = fmt.Errorf("无法识别上传文件的格式,目前支持的文件格式:\n%v\n请将相片或视频转换成支持的格式再上传", strings.Join(fileExtensions, " "))
		renderError(w, err.Error(), http.StatusUnsupportedMediaType)
		logrus.WithError(err).Errorf("unknown media type, uloaded filename:%s", fileHeader.Filename)
		return err
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == filetype.Unknown {
		err = fmt.Errorf("filetype.Match: cannot file a matched file type")
		logrus.WithError(err).Errorf("unknown media type, uloaded filename:%s", fileHeader.Filename)
		return err
	}

	fileName := randToken(12)

	newPath := filepath.Join(s.receiveDir, fileName+"."+kind.Extension)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		renderError(w, fmt.Sprintf("create file failed:%v", err), http.StatusInternalServerError)
		logrus.WithError(err).Error("create file failed")
		return err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(w, "Write file failed", http.StatusInternalServerError)
		logrus.WithError(err).Error("Write file failed")
		return err
	}
	logrus.WithFields(logrus.Fields{"origin filename": fileHeader.Filename, "newPath": newPath}).Infof("Save uploaded file")
	return nil
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

			if err := uploadFile(w, r); err != nil {
				logrus.WithError(err).Errorf("Upload file failed.")
				return
			}
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
