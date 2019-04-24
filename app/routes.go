package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/jusongchen/lepus/version"
	"github.com/sirupsen/logrus"
)

type profileCtxKeyType string

const profileCtxKey profileCtxKeyType = "profile-json"

// ErrNoAlumnusName is a customized error
var ErrNoAlumnusName = errors.New("missing context value for profile-json")

// GetAlumnusProfile gets alumnus profile from context
// returns ErrNoAlumnusProfile if there is no AlumnusProfile.
func getAlumnusProfile(ctx context.Context) string {
	// func getAlumnusProfile(ctx context.Context) (string, error) {

	fmt.Printf("ctxValue:%+v\n", ctx.Value(profileCtxKey))

	// profile, ok := ctx.Value(profileCtxKey).(AlumnusProfile)
	profileJSON, ok := ctx.Value(profileCtxKey).(string)
	if !ok {
		return ""
	}

	return profileJSON
}

// AddProfile adds user profile
func AddProfile(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// if we already have a profile
		if profileJSON := getAlumnusProfile(r.Context()); profileJSON != "" {
			next.ServeHTTP(w, r)
			return
		}

		// now try to get profile from Form
		profile, err := newUserProfileFromForm(w, r)
		if err != nil || profile == nil {
			logrus.WithError(err).Debugf("cannot find profile at URL:%v", r.URL)

			next.ServeHTTP(w, r)
			return
		}

		logrus.Infof("Get New Profile from form:%v", profile)

		b, err := json.Marshal(profile)
		if err != nil {
			logrus.WithError(err).Errorf("json Marshall failed:%v", profile)
		}

		ctx := context.WithValue(r.Context(), profileCtxKey, string(b))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
	// logger.Formatter = &logrus.JSONFormatter{
	logger.Formatter = &logrus.TextFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}
	r := s.router
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(AddProfile)

	registerStaticWeb(r, filepath.Join(s.lepusHomeDir, staticDir))
	r.HandleFunc("/home", s.handlerHome())
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		s.notFound(w)
	})

	r.Handle("/selectphoto", s.selectPhotoHandler())
	r.Post("/where2", s.where2Handler())
	r.Post("/sponsor", s.sponsorHandler())

	r.Handle("/signup", s.signupHandler())

}

func (s *lepus) signupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s.Render(w, "signup", s.educatorNames)

		case "POST":

			r.Method = "GET"
			ctx := r.Context()
			fmt.Printf("ctxValue:%+v\n", ctx.Value(profileCtxKey))

			use gorilla session 
			http.Cookie(Name:"user_profile",Value:ctx.Value(profileCtxKey),)
			http.Redirect(w, r, "/selectphoto", http.StatusSeeOther)

		default:
			fmt.Fprintf(w, "Unknown http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) selectPhotoHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error

		switch r.Method {
		case "GET":
			profileJSON := getAlumnusProfile(r.Context())
			if profileJSON == "" {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：cannot get AlumnusProfile from http request:%v", r.URL))
				return
			}

			profile, err := getParticipantProfile(profileJSON)
			if err != nil {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：getParticipantProfile failed at:%v", r.URL))
				return
			}

			data := struct {
				EducatorNames []string
				SessionID     string
			}{
				EducatorNames: profile.SelectedEducators,
				SessionID:     profileJSON,
			}
			s.Render(w, "selectphoto", data)

		// case "POST":

		// 	if err = r.ParseForm(); err != nil {
		// 		s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT_PARSE_FORM at URL:%v", r.URL))
		// 		return
		// 	}
		// 	sessionID, err := getSessionID(w, r)
		// 	if err != nil {
		// 		s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT_GET_SESSION_ID at URL:%v", r.URL))
		// 		return
		// 	}

		// 	profile, err := getParticipantProfile(sessionID)
		// 	if err != nil {
		// 		s.serverErrorWithMsg(w, err, fmt.Sprintf("内部错误：CANNOT get participant profile at URL:%v", r.URL))
		// 		return
		// 	}

		// 	logrus.WithField("URL", r.URL).Infof("Participant Profile:%+v", profile)

		// 	data := struct {
		// 		EducatorNames []string
		// 		SessionID     string
		// 	}{
		// 		EducatorNames: profile.SelectedEducators,
		// 		SessionID:     sessionID,
		// 	}
		// 	s.Render(w, "selectphoto", data)
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
			imageFilePath := ""
			if imageFile != "" {
				imageFilePath = s.imageDir + "/" + imageFile
			}
			if err != nil {
				data = ctrlDataTyp{
					SessionID: sessionID,
					InfoText:  err.Error(),
					InfoType:  "danger",
					ImageFile: imageFilePath,
				}
			} else {
				data = ctrlDataTyp{
					SessionID: sessionID,
					InfoText:  "上传成功",
					InfoType:  "success",
					ImageFile: imageFilePath,
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
