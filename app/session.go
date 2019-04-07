package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type sessionIDType struct {
	Name              string
	GradYear          string
	SelectedEducators []string
}

func (sid *sessionIDType) JSONMarshal() string {

	b, err := json.Marshal(sid)
	if err != nil {
		msg := fmt.Sprintf("json Marshall %v failed:%v", sid, err)
		log.Fatalf(msg)
	}

	return string(b)

}

func getSessionID(w http.ResponseWriter, r *http.Request) sessionIDType {

	sessionID := sessionIDType{
		Name:              r.Form["name"][0],
		GradYear:          r.Form["gradYear"][0],
		SelectedEducators: r.Form["educators"],
	}

	return sessionID

}
