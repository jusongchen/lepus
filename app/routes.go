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

	"github.com/disintegration/imaging"
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

func resizeImage(srcFile, dstFile string) error {

	srcImage, err := imaging.Open(srcFile)

	if err != nil {
		logrus.WithError(err).Error("resizeImage():open image failed")
		return err
	}

	dstImage128 := imaging.Resize(srcImage, 128, 128, imaging.Lanczos)

	err = imaging.Save(dstImage128, dstFile)
	if err != nil {
		logrus.WithError(err).Error("resizeImage():save image failed")
		return err
	}
	return nil

}

//uploadFile returns resized image file and error
func uploadFile(w http.ResponseWriter, r *http.Request) (string, error) {

	var err error
	// validate file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err = r.ParseMultipartForm(maxUploadSize)

	if err != nil {

		// renderError(w, errMsg, http.StatusBadRequest)
		return "", fmt.Errorf("上传的文件太大（已超过%d兆字节）", maxUploadSize/1024/1024)
	}

	fmt.Printf("upload photo form value:%s\n", r.Form)

	// parse and validate file and post parameters
	// fileType := r.PostFormValue("type")

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		return "", fmt.Errorf("内部错误: r.FromFile get error:%s", err)
	}

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("内部错误，无法读取上传文件:%v", err)
	}

	fileExtensions := []string{"jpg", "png", "gif", "webp", "cr2", "tif", "bmp", "heif", "jxr", "psd", "ico", "mp4", "m4v", "mkv", "webm", "mov", "avi", "wmv", "mpg", "flv"}

	if !(filetype.IsImage(fileBytes) || filetype.IsVideo(fileBytes) || filetype.IsAudio(fileBytes)) {
		return "", fmt.Errorf("无法识别上传文件的格式,目前支持的文件格式:\n%v\n请将相片或视频转换成支持的格式再上传", strings.Join(fileExtensions, " "))
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == filetype.Unknown {
		return "", fmt.Errorf("无法识别上传文件的格式")
	}

	fileName := randToken(12) + "." + kind.Extension

	newPath := filepath.Join(s.receiveDir, fileName)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		return "", fmt.Errorf("内部错误:create file failed:%v", err)
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		return "", fmt.Errorf("内部错误:Write file failed:%v", err)
	}
	logrus.WithFields(logrus.Fields{"origin filename": fileHeader.Filename, "newPath": newPath}).Infof("Save uploaded file")

	imagePathRelative := filepath.Join(s.imageDir, fileName)

	if err = resizeImage(newPath, filepath.Join(s.staticHomeDir, imagePathRelative)); err != nil {
		// just log error, we may get an error during resize the picture as we do not handle all formats
		logrus.WithError(err).WithField("filename", newPath).Error("resize image failed")
		//do not return error here
		return "", nil
	}

	return imagePathRelative, nil
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

	type ctrlDataTyp struct {
		SessionID     string
		InfoText      string
		InfoType      string
		ImageFile     string
		ShowImageFile bool
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		// case "GET":
		case "POST":

			imageFile, err := uploadFile(w, r)
			if err != nil {
				logrus.WithError(err).Errorf("uploadFile failed")
			}

			sessionID := s.getSessionID(w, r)
			data := ctrlDataTyp{SessionID: sessionID,
				InfoText:      "上传成功",
				InfoType:      "success",
				ImageFile:     imageFile,
				ShowImageFile: imageFile != "",
			}

			if err != nil {
				data.InfoText = err.Error()
				data.InfoType = "danger"
			}
			logrus.Infof("rendering where2 with data %v", data)

			s.Render(w, "where2", data)
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
