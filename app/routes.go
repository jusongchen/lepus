package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/jusongchen/lepus/version"
	log "github.com/sirupsen/logrus"
)

const lepusSessionName = "alumnus_profile"
const eduSessionValKey = "educator_names"
const nameSessionValKey = "alumnus_name"
const gradYearSessionValKey = "alumnus_gradyear"
const imagFileSessionValKey = "resizedFilename"
const uploadInfoTextValKey = "infoText"
const uploadInfoTypeValKey = "infoType"

func (s *lepus) routes(staticDir string) {

	// A good base middleware stack
	logger := log.New()
	// logger.Formatter = &log.JSONFormatter{
	logger.Formatter = &log.TextFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}
	r := s.router
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	registerStaticWeb(r, filepath.Join(s.lepusHomeDir, staticDir))
	r.HandleFunc("/home", s.handlerHome())
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		s.notFound(w)
	})

	r.Handle("/selectphoto", s.selectPhotoHandler())
	r.Handle("/where2", s.where2Handler())
	r.Handle("/sponsor", s.sponsorHandler())
	r.Handle("/signup", s.signupHandler())
	r.Handle("/listmedia", s.listmediaHandler())

}

func (s *lepus) signupHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s.Render(w, "signup", s.educatorNames)

		case "POST":

			// now try to get profile from Form
			profile, err := newUserProfileFromForm(w, r)
			if err != nil || profile == nil {
				log.WithError(err).Debugf("cannot find profile at URL:%v", r.URL)
				s.serverErrorWithMsg(w, err, "Internal Error:Cannot get User Profile")
				return
			}

			session, _ := s.cookieStore.Get(r, lepusSessionName)
			// Set some session values.
			session.Values[nameSessionValKey] = profile.Name
			session.Values[gradYearSessionValKey] = profile.GradYear
			session.Values[eduSessionValKey] = profile.SelectedEducators

			// Save it before we write to the response/return from the handler.
			session.Save(r, w)
			// r.Method = "GET"
			http.Redirect(w, r, "/selectphoto", http.StatusSeeOther)

		default:
			fmt.Fprintf(w, "Not handled http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) getUserProfile(w http.ResponseWriter, r *http.Request) *AlumnusProfile {

	session, _ := s.cookieStore.Get(r, lepusSessionName)

	name, ok := session.Values[nameSessionValKey].(string)
	if !ok {
		s.serverError(w, errors.New("Expect name values in cookie but not found"))
		return nil
	}

	gradyear, ok := session.Values[gradYearSessionValKey].(string)
	if !ok {
		s.serverError(w, errors.New("Expect gradyear values in cookie but not found"))
		return nil
	}
	educatorNames, ok := session.Values[eduSessionValKey].([]string)
	if !ok {
		s.serverError(w, errors.New("Expect educators values in cookie but not found"))
		return nil
	}
	return &AlumnusProfile{
		Alumnus: Alumnus{
			Name:     name,
			GradYear: gradyear,
		},
		SelectedEducators: educatorNames,
	}
}

func (s *lepus) selectPhotoHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error

		switch r.Method {
		case "GET":

			profile := s.getUserProfile(w, r)
			if profile == nil {
				return
			}

			data := struct{ EducatorNames []string }{EducatorNames: profile.SelectedEducators}
			s.Render(w, "selectphoto", data)

		case "POST":

			var err error

			profile := s.getUserProfile(w, r)
			if profile == nil {
				s.serverErrorWithMsg(w, err, fmt.Sprintf("cannot get user profile at %v", r.URL))
				return
			}

			// resizedFilename is a filename within public/images folder
			rpt, err := s.uploadFile(w, r)

			//file upload succeeded
			if err == nil {
				rpt.AlumnusProfile = *profile
				rpt.forEducators = r.Form["educators"]
				if rpt.MediaType == imageMedia {
					fileName := rpt.saveAsName
					newPath := filepath.Join(s.receiveDir, fileName)

					if err = resizeImage(newPath, filepath.Join(s.staticHomeDir, s.imageDir, fileName)); err != nil {
						// just log error, we may get an error during resize the picture as we do not handle all formats
						log.WithError(err).WithField("filename", newPath).Error("resize image failed")
						//do not return error here, as even resize failed, we still move forward
					} else {
						rpt.resizedFilename = fileName
					}
				}

				err1 := s.SaveUpload(rpt)
				if err1 != nil {
					s.serverErrorWithMsg(w, err1, fmt.Sprintf("Internal DB Error"))
					// continue to ?
				}
			}

			resizedFilename := rpt.resizedFilename

			infoText := "上传成功!"
			infoType := "success"
			if err != nil {
				infoText = err.Error()
				infoType = "danger"
			}

			if rpt.FileSize != 0 {
				infoText += fmt.Sprintf(" 文件大小:%.2f MB 上传用时:%v", float64(rpt.FileSize)/1024/1024, rpt.Duration)
			}

			if resizedFilename != "" {
				resizedFilename = s.imageDir + "/" + resizedFilename
			}

			session, _ := s.cookieStore.Get(r, lepusSessionName)

			session.Values[imagFileSessionValKey] = resizedFilename
			session.Values[uploadInfoTextValKey] = infoText
			session.Values[uploadInfoTypeValKey] = infoType

			session.Save(r, w)
			http.Redirect(w, r, "/where2", http.StatusSeeOther)

		default:
			err = fmt.Errorf("BadRequest")
			s.serverErrorWithMsg(w, err, fmt.Sprintf("Not handled http method for url %s:%s", r.URL, r.Method))
			return
		}
	})
}

func (s *lepus) where2Handler() http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":

			session, _ := s.cookieStore.Get(r, lepusSessionName)

			infoText, ok := session.Values[uploadInfoTextValKey].(string)
			if !ok {
				s.serverError(w, errors.New("Expect uploadInfoTextValKey values in cookie but not found"))
				return
			}
			infoType, ok := session.Values[uploadInfoTypeValKey].(string)
			if !ok {
				s.serverError(w, errors.New("Expect uploadInfoTypeValKey values in cookie but not found"))
				return
			}
			resizedFilename, ok := session.Values[imagFileSessionValKey].(string)
			if !ok {
				s.serverError(w, errors.New("Expect imagFileSessionValKey values in cookie but not found"))
				return
			}

			data := struct {
				InfoText    string
				InfoType    string
				ResizedFile string
			}{
				InfoText:    infoText,
				InfoType:    infoType,
				ResizedFile: resizedFilename,
			}
			s.Render(w, "where2", data)

		default:
			fmt.Fprintf(w, "Not handled http method for url %s:%s", r.URL, r.Method)
		}
	})
}

func (s *lepus) sponsorHandler() http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			s.Render(w, "sponsor", struct{}{})
		default:
			fmt.Fprintf(w, "Not handled http method for url %s:%s", r.URL, r.Method)
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
			log.WithError(err).Errorf("Could not encode data:%v", info)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}

func (s *lepus) listmediaHandler() http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			to := time.Now()
			//TODO, support time ranged search
			from := to
			m, err := s.getUploadedMedia(from, to)
			if err != nil {
				s.serverError(w, err)
			}

			data := struct {
				MediaFiles []Media
			}{
				MediaFiles: m,
			}

			s.Render(w, "listmedia", data)
		default:
			fmt.Fprintf(w, "Not handled http method for url %s:%s", r.URL, r.Method)
		}
	})
}
