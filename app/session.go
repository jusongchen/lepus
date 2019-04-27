package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getParticipantProfile(sessionID string) (AlumnusProfile, error) {

	profile := AlumnusProfile{}
	err := json.Unmarshal([]byte(sessionID), &profile)
	if err != nil {
		log.Errorf(`Failed to unmarshal to participantProfile:
			input:%s
			Struct:%+v
			err:%v`, sessionID, profile, err)
	}
	return profile, err

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
