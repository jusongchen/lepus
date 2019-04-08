package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type participantProfileTyp struct {
	Name              string
	GradYear          string
	SelectedEducators []string
}

func getParticipantProfile(sessionID string) (participantProfileTyp, error) {

	profile := participantProfileTyp{}
	err := json.Unmarshal([]byte(sessionID), &profile)
	if err != nil {
		log.Fatalf(`Failed to unmarshal to participantProfile:
			input:%s
			Struct:%+v
			err:%v`, sessionID, profile, err)
	}
	return profile, err

}

func (s *lepus) getSessionID(w http.ResponseWriter, r *http.Request) (sessionID string) {
	// two cases:
	// 		1) form post from /where2. in this case, the sessionID string is set
	//		2) form post from /signup. in this case , the sessionID string is not set, but name,gradYear,educators are set

	defer func() {
		log.Printf("sessionID at %s:%+v", r.URL, sessionID)
	}()

	sID := r.Form["sessionID"]
	if sID != nil {
		sessionID = sID[0]
		return
	}
	sessionID = s.newSessionIDFromForm(w, r)
	return
}

func (s *lepus) newSessionIDFromForm(w http.ResponseWriter, r *http.Request) string {

	// session not found
	if r.Form["name"] == nil || r.Form["gradYear"] == nil {
		errMsg := fmt.Sprintf("sessionID missing,  and missing input name or gradYear, URL:%v", r.URL)
		renderError(w, errMsg, http.StatusInternalServerError)
		log.Fatalf(errMsg)
		return ""
	}

	profile := participantProfileTyp{
		Name:              r.Form["name"][0],
		GradYear:          r.Form["gradYear"][0],
		SelectedEducators: r.Form["educators"],
	}

	b, err := json.Marshal(profile)
	if err != nil {
		errMsg := fmt.Sprintf("json Marshall %v failed:%v", profile, err)
		renderError(w, errMsg, http.StatusInternalServerError)
		log.Fatalf(errMsg)
		return ""
	}

	return string(b)

}
