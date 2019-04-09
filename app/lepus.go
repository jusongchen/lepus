// copyright Jusong Chen

package app

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"unicode/utf8"

	"github.com/go-chi/chi"
	"github.com/jusongchen/lepus/chn"
)

const maxUploadSize = 20 * 1024 * 1024 // 20 mb

//lepus is the server to implement all Lepus features
type lepus struct {
	router        *chi.Mux
	addr          string
	version       string
	receiveDir    string
	educatorNames []string
	viewPath      string
	//tempMap key-> urlPath , such as signUp
	tempMap map[string]*template.Template
	httpSrv *http.Server
}

var s lepus

//Start starts Lepus server
func Start(addr string, staticHomeDir string, srvVersion string, receiveDir string, educatorNames []string, viewPath string) {

	// insert an whitespace if educatorNames is less than 2 charactor long

	for i, name := range educatorNames {
		if utf8.RuneCountInString(name) == 2 {
			educatorNames[i] = string([]rune(name)[0]) + "　" + string([]rune(name)[1])
		}
	}
	// sort educatorNames

	sort.Sort(chn.ByPinyin(educatorNames))

	s = lepus{
		router:        chi.NewRouter(),
		addr:          addr,
		version:       srvVersion,
		receiveDir:    receiveDir,
		educatorNames: educatorNames,
		viewPath:      viewPath,
	}
	s.initTemplates(viewPath)
	s.routes(staticHomeDir)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}

	// this channel is for graceful shutdown:
	// if we receive an error, we can send it here to notify the LepusServer to be stopped
	shutdown := make(chan struct{}, 1)
	go func() {
		err := s.httpSrv.ListenAndServe()
		if err != nil {
			shutdown <- struct{}{}
			log.Printf("%v", err)
		}
	}()
	log.Printf("The service is ready to listen and serve on %s.", s.httpSrv.Addr)

	select {
	case killSignal := <-interrupt:
		switch killSignal {
		case os.Interrupt:
			log.Print("Got SIGINT...")
		case syscall.SIGTERM:
			log.Print("Got SIGTERM...")
		}
	case <-shutdown:
		log.Printf("Get server shutdown request")
	}

	log.Print("The service is shutting down...")
	s.httpSrv.Shutdown(context.Background())
	log.Print("Done")
}

//Stop will stop the  Lepus app
func Stop() {
	s.httpSrv.Shutdown(context.Background())

}
