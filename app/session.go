package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func getParticipantProfile(sessionID string) (AlumnusProfile, error) {

	profile := AlumnusProfile{}
	err := json.Unmarshal([]byte(sessionID), &profile)
	if err != nil {
		logrus.Errorf(`Failed to unmarshal to participantProfile:
			input:%s
			Struct:%+v
			err:%v`, sessionID, profile, err)
	}
	return profile, err

}

func (s *lepus) getSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	// two cases:
	// 		1) form post from /where2. in this case, the sessionID string is set
	//		2) form post from /signup. in this case , the sessionID string is not set, but name,gradYear,educators are set

	sID := r.Form["sessionID"]
	if sID != nil {
		return sID[0], nil
	}
	return s.newSessionIDFromForm(w, r)
}

func (s *lepus) newSessionIDFromForm(w http.ResponseWriter, r *http.Request) (string, error) {

	// session not found
	if r.Form["name"] == nil || r.Form["gradYear"] == nil {
		return "", fmt.Errorf("Cannot get sessionID for URL:%v", r.URL)
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
		return "", err
	}
	b, err := json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("json Marshall %v failed:%v", profile, err)
	}

	return string(b), nil

}
