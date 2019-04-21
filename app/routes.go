package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
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
		s.notFound(w)
	})

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

		var err error

		switch r.Method {
		// case "GET":
		case "POST":

			if err = r.ParseForm(); err != nil {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT_PARSE_FORM at URL:%v", r.URL))
				return
			}
			sessionID, err := s.getSessionID(w, r)
			if err != nil {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT_GET_SESSION_ID at URL:%v", r.URL))
				return
			}

			profile, err := getParticipantProfile(sessionID)
			if err != nil {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT get participant profile at URL:%v", r.URL))
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
			s.serverErrorWithMsg(w, err, fmt.Sprintf("Unknown http method for url %s:%s", r.URL, r.Method))
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
