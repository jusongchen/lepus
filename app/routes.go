package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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
	r.Handle("/listmedia", s.listMediaHandler())
	r.Handle("/exportmedia", s.exportMediaHandler())

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

func newUserProfileFromForm(w http.ResponseWriter, r *http.Request) (*AlumnusProfile, error) {
	r.ParseForm()
	log.Errorf("form :%+v\n", r.Form)
	// session not found
	if r.Form["name"] == nil || r.Form["gradYear"] == nil || r.Form["educators"] == nil {
		return nil, fmt.Errorf("Cannot get sessionID for URL:%v", r.URL)
	}

	profile := AlumnusProfile{
		Alumnus: Alumnus{
			Name:     r.Form["name"][0],
			GradYear: r.Form["gradYear"][0],
		},
		SelectedEducators: r.Form["educators"],
	}

	_, err := s.SaveSignup(profile)
	if err != nil {
		log.WithError(err).WithField("profile", profile).Error("Save profile to DB failed")
	}
	return &profile, err

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

					if err = resizeImage(bytes.NewReader(rpt.filedata), filepath.Join(s.staticHomeDir, s.imageDir, fileName)); err != nil {
						// just log error, we may get an error during resize the picture as we do not handle all formats
						log.WithError(err).Error("resize image failed")
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
				infoText += fmt.Sprintf(" 大小:%.2f MB 用时:%.2f秒", float64(rpt.FileSize)/1024/1024, rpt.Duration.Seconds())
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
			version.BuildTime, version.Commit, version.Release, "Developed by Jusong Chen 陈居松",
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

func (s *lepus) listMediaHandler() http.HandlerFunc {

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

func saveMediaFile(exp2Dir, educatorName, alumnusName, gradYear, originFileExt string, fileBytes []byte) error {

	eduPath := filepath.Join(exp2Dir, educatorName)
	if _, err := os.Stat(eduPath); os.IsNotExist(err) {
		err1 := os.Mkdir(eduPath, 0700)
		if err1 != nil {
			err = fmt.Errorf("create export dir for Edu %s failed:%v", educatorName, err)
			return err
		}
	}

	defaultPath := filepath.Join(eduPath, gradYear+"_"+alumnusName)
	save2Path := defaultPath
	for _, postfix := range []string{"", "_1", "_2", "_3", "_4", "_5", "_6", "_7", "_8", "_9", "_z"} {
		save2Path = defaultPath + postfix
		//check if the filename has already been used
		if _, err := os.Stat(save2Path); os.IsNotExist(err) {
			break
		}
		if postfix == "_z" {
			return fmt.Errorf("Too many duplicated uploads for :%v", save2Path)
		}
	}

	// write file
	newFile, err := os.Create(save2Path + originFileExt)
	if err != nil {
		err = fmt.Errorf("内部错误:create file failed:%v", err)
		return err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		err = fmt.Errorf("内部错误:Write file failed:%v", err)
		return err
	}
	log.Infof("media file saved:%s", save2Path)

	return nil
}

func (s *lepus) exportMediaHandler() http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case "GET":

			//export to dir
			exp2dir := filepath.Join(s.export2Dir, time.Now().Format("2006-01-02T15-04-05"))
			if _, err := os.Stat(exp2dir); os.IsNotExist(err) {
				err1 := os.MkdirAll(exp2dir, 0700)
				if err1 != nil {
					s.serverErrorWithMsg(w, err1, "create export dir failed")
				}

			}
			to := time.Now()
			//TODO, support time ranged search
			from := to
			media, err := s.getUploadedMedia(from, to)
			if err != nil {
				s.serverError(w, err)
			}

			stat := struct {
				TotalUploads   int
				EduCount       map[string]int
				GradYearCount  map[string]int
				ExportAttempts int
				ExportFails    int
			}{
				EduCount:      map[string]int{},
				GradYearCount: map[string]int{},
			}

			for _, m := range media {
				//get media data
				filedata, err := s.getMediaDataByID(m.MediaID)
				if err != nil {
					s.serverError(w, err)
					return
				}
				stat.GradYearCount[m.AlumnusGradYear]++
				stat.TotalUploads++

				originFileExt := filepath.Ext(m.SaveAsName)
				//save to each Educator's folder
				for _, eduName := range m.ForEducators {
					stat.ExportAttempts++
					stat.EduCount[eduName]++
					err := saveMediaFile(exp2dir, eduName, m.AlumnusName, m.AlumnusGradYear, originFileExt, filedata)
					if err != nil {
						stat.ExportFails++
						s.serverErrorWithMsg(w, err, fmt.Sprintf(`save media file failed: exp2dir:%v eduName:%v AlumnusName:%v AlumnusGradYear:%v fileExt:%s`, exp2dir, eduName, m.AlumnusName, m.AlumnusGradYear, originFileExt))
					}
				}

			}
			absPath, _ := filepath.Abs(exp2dir)
			asJSON, err2 := json.Marshal(stat)
			if err2 != nil {
				s.serverErrorWithMsg(w, err2, fmt.Sprintf(`Json marshal failed: %+v`, stat))

			}
			msg := fmt.Sprintf("export to %s\n stat: %s", absPath, asJSON)
			http.Error(w, msg, http.StatusOK)
		default:
			fmt.Fprintf(w, "Not handled http method for url %s:%s", r.URL, r.Method)
		}
	})
}
