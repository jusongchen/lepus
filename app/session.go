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

func (sid participantProfileTyp) JSONMarshal() string {

	b, err := json.Marshal(sid)
	if err != nil {
		msg := fmt.Sprintf("json Marshall %v failed:%v", sid, err)
		log.Fatalf(msg)
	}

	return string(b)

}

func getParticipantProfile(w http.ResponseWriter, r *http.Request) participantProfileTyp {

	// two cases:
	// 		1) form post from /where2. in this case, the sessionID string is set
	//		2) form post from /signup. in this case , the sessionID string is not set, but name,gradYear,educators are set
	// called from /where2
	//

	sessionID := ""
	if r.Form["sessionID"] != nil {
		sessionID = r.Form["sessionID"][0]
	}

	profile := participantProfileTyp{}

	//  sessionID:{"Name":"JUSO","GradYear":"33","SelectedEducators":["陈由溪","蒋永潮","林昭英","潘世英"]}
	if sessionID != "" {
		if err := json.Unmarshal([]byte(sessionID), &profile); err != nil {
			log.Fatalf(`Failed to unmarshal to participantProfile:
			input:%s
			Struct:%+v
			err:%v`, sessionID, profile, err)
		}
		return profile
	}

	if r.Form["name"] == nil {
		log.Fatalf("missing input name, URL:%v", r.URL)
		return profile
	}

	if r.Form["gradYear"] == nil {
		log.Fatalf("missing input gradYear, URL:%v", r.URL)
		return profile
	}

	return participantProfileTyp{
		Name:              r.Form["name"][0],
		GradYear:          r.Form["gradYear"][0],
		SelectedEducators: r.Form["educators"],
	}

}
