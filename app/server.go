package app

import (
	"context"
	// "database/sql"
	sql "github.com/jmoiron/sqlx"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sort"
	"syscall"
	"unicode/utf8"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/jusongchen/lepus/chn"
	log "github.com/sirupsen/logrus"
)

const maxUploadSize = 200 * 1024 * 1024 //MB

//lepus is the server to implement all Lepus features
type lepus struct {
	router        *chi.Mux
	addr          string
	errorLog      *log.Logger
	lepusHomeDir  string
	receiveDir    string
	staticHomeDir string
	imageDir      string
	educatorNames []string
	viewPath      string
	tempMap       map[string]*template.Template
	httpSrv       *http.Server
	store         *dbStore
	cookieStore   *sessions.CookieStore
}

var s lepus

//Start starts Lepus server
func Start(db *sql.DB, sessionKey, addr, lepusHomeDir, staticHomeDir, receiveDir, imageDir, viewPath string, educatorNames []string) {

	errorLog := log.New()
	// .New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	s = lepus{
		router:        chi.NewRouter(),
		addr:          addr,
		lepusHomeDir:  lepusHomeDir,
		receiveDir:    receiveDir,
		staticHomeDir: staticHomeDir,
		imageDir:      imageDir,
		viewPath:      viewPath,
		errorLog:      errorLog,
		cookieStore:   sessions.NewCookieStore([]byte(sessionKey)),
	}

	// insert an whitespace if educatorNames is less than 2 charactor long
	for i, name := range educatorNames {
		if utf8.RuneCountInString(name) == 2 {
			educatorNames[i] = string([]rune(name)[0]) + "　" + string([]rune(name)[1])
		}
	}
	// sort educatorNames
	sort.Sort(chn.ByPinyin(educatorNames))
	s.educatorNames = educatorNames

	for _, dir := range []string{filepath.Join(s.staticHomeDir, s.imageDir), s.receiveDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err1 := os.Mkdir(dir, 0700)
			if err1 != nil {
				log.Fatalf("Create dir '%s' failed:%s", dir, err1)
			}
		}
	}

	if err := initSqliteStore(db); err != nil {
		return
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
}

//Stop will stop the  Lepus app
func Stop() {
	s.httpSrv.Shutdown(context.Background())

}

func (s *lepus) serverErrorWithMsg(w http.ResponseWriter, err error, msg string) {
	s.serverError(w, errors.WithMessage(err, msg))
}

//serverError logs err, then sends a generic 500 Internal Server Error response to the user.
func (s *lepus) serverError(w http.ResponseWriter, err error) {
	s.errorLog.Error(err.Error() + "\n" + string(debug.Stack()))

	http.Error(w, "内部错误:"+err.Error(), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user
func (s *lepus) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// not found
func (s *lepus) notFound(w http.ResponseWriter) {
	s.clientError(w, http.StatusNotFound)
}
