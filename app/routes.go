package app

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-chi/chi/middleware"
	"github.com/h2non/filetype"
	"github.com/jusongchen/lepus/version"
	"github.com/sirupsen/logrus"
)

/*
	routing path:
	at home /
		post -> /signup
			post -> /selectphoto
				post ->  /where2
					post 1) /home
						 2) /selectphoto
						 3) /sponsor
							post -> /home
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

	registerStaticWeb(s.router, filepath.Join(s.lepusHomeDir, staticDir))

	s.router.Handle("/signup", s.signupHandler())
	s.router.Handle("/selectphoto", s.selectPhotoHandler())
	s.router.Post("/where2", s.where2Handler())
	s.router.Post("/sponsor", s.sponsorHandler())
	s.router.HandleFunc("/home", s.handlerHome())
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		renderError(w, fmt.Sprintf("path not found:%s", r.URL), http.StatusNotFound)
	})

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
	logrus.Infof("resizeImage():save image file to %s", dstFile)
	return nil

}

//uploadFile returns resized image file and error
func (s *lepus) uploadFile(w http.ResponseWriter, r *http.Request) (sessionID string, resized_filename string, err error) {

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err = r.ParseMultipartForm(maxUploadSize)

	if err != nil {
		// renderError(w, errMsg, http.StatusBadRequest)
		err = fmt.Errorf("上传的文件太大（已超过%d兆字节）:%v", maxUploadSize/1024/1024, err)
		return
	}

	sessionID, err = s.getSessionID(w, r)
	if err != nil {
		logrus.Errorf("内部错误：CANNOT_GET_SESSION_ID at URL:%v", r.URL)
	} else {
		logrus.Infof("get SessionID %s at URL:%v", sessionID, r.URL)
	}

	// fmt.Printf("upload photo form value:%s\n", r.Form)

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return
	}

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return
	}

	fileExtensions := []string{"jpg", "png", "gif", "webp", "cr2", "tif", "bmp", "heif", "jxr", "psd", "ico", "mp4", "m4v", "mkv", "webm", "mov", "avi", "wmv", "mpg", "flv"}

	isImage := filetype.IsImage(fileBytes)
	if !(isImage || filetype.IsVideo(fileBytes) || filetype.IsAudio(fileBytes)) {
		err = fmt.Errorf("无法识别上传文件的格式,目前支持的文件格式:\n%v\n请将相片或视频转换成支持的格式再上传", strings.Join(fileExtensions, " "))
		return
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == filetype.Unknown {
		err = fmt.Errorf("无法识别上传文件的格式")
		return
	}

	fileName := randToken(12) + "." + kind.Extension

	newPath := filepath.Join(s.receiveDir, fileName)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		err = fmt.Errorf("内部错误:create file failed:%v", err)
		return
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		err = fmt.Errorf("内部错误:Write file failed:%v", err)
		return
	}
	logrus.WithFields(logrus.Fields{"originFilename": fileHeader.Filename, "newPath": newPath}).Infof("Save uploaded file")

	if !isImage {
		return
	}

	if err = resizeImage(newPath, filepath.Join(s.staticHomeDir, s.imageDir, fileName)); err != nil {
		// just log error, we may get an error during resize the picture as we do not handle all formats
		logrus.WithError(err).WithField("filename", newPath).Error("resize image failed")
		//do not return error here, as even resize failed, we still move forward
		err = nil
		resized_filename = ""
		return
	}
	resized_filename = fileName
	return
}

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

		var errMsg string
		var err error
		defer func() {
			if err != nil {
				renderError(w, errMsg, http.StatusBadRequest)
				logrus.WithError(err).Error(errMsg)
			}
		}()

		switch r.Method {
		// case "GET":
		case "POST":

			if err = r.ParseForm(); err != nil {
				errMsg = fmt.Sprintf("内部错误：CANNOT_PARSE_FORM at URL:%v", r.URL)
				return
			}
			sessionID, err := s.getSessionID(w, r)
			if err != nil {
				errMsg = fmt.Sprintf("内部错误：CANNOT_GET_SESSION_ID at URL:%v", r.URL)
				return
			}

			profile, err := getParticipantProfile(sessionID)
			if err != nil {
				errMsg = fmt.Sprintf("内部错误：CANNOT get participant profile at URL:%v", r.URL)
				return
			}

			logrus.WithField("URL", r.URL).Infof("Participant Profile:%+v", profile)

			data := struct {
				EducatorNames []string
				SessionID     string
			}{
				EducatorNames: profile.SelectedEducators,
				SessionID:     sessionID,
			}
			s.Render(w, "selectphoto", data)
		default:
			err = fmt.Errorf("BadRequest")
			errMsg = fmt.Sprintf("Unknown http method for url %s:%s", r.URL, r.Method)
			return
		}
	})
}

func (s *lepus) where2Handler() http.HandlerFunc {

	type ctrlDataTyp struct {
		SessionID string
		InfoText  string
		InfoType  string
		ImageFile string
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		// case "GET":
		case "POST":
			var sessionID, imageFile string
			var err error
			data := ctrlDataTyp{}

			// imageFile is a filename within public/images folder
			sessionID, imageFile, err = s.uploadFile(w, r)
			if err != nil {
				data = ctrlDataTyp{
					SessionID: sessionID,
					InfoText:  err.Error(),
					InfoType:  "danger",
					ImageFile: s.imageDir + "/" + imageFile,
				}
			} else {
				data = ctrlDataTyp{
					SessionID: sessionID,
					InfoText:  "上传成功",
					InfoType:  "success",
					ImageFile: s.imageDir + "/" + imageFile,
				}
			}

			logrus.Infof("rendering where2 with data %+v", data)

			s.Render(w, "where2", data)
		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) sponsorHandler() http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		// case "GET":
		case "POST":
			s.Render(w, "sponsor", struct{}{})
		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

// handlerHome returns a simple HTTP handler function which writes a response.
func (s *lepus) handlerHome() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		info := struct {
			BuildTime string `json:"buildTime"`
			Commit    string `json:"commit"`
			Release   string `json:"release"`
			Message   string `json:"message"`
		}{
			version.BuildTime, version.Commit, version.Release, "Hello Lepus Administrators!",
		}

		body, err := json.Marshal(info)
		if err != nil {
			logrus.WithError(err).Errorf("Could not encode data:%v", info)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}
