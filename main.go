// copyright Jusong Chen

package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

const version = "0.2"
const maxUploadSize = 20 * 1024 * 1024 // 20 mb
const defaultPort = "8080"
const dirForPhotos = "photos"
const dirForStatic = "./public"
const registrationHTMLFilename = "registration.html"

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func main() {
	fmt.Printf("Lepus version:%s", version)

	if _, err := os.Stat(dirForStatic); os.IsNotExist(err) {
		log.Fatalf("Directory for static web content does not exist:%s", dirForStatic)

	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s -port <port number>\n", os.Args[0])
		flag.PrintDefaults()
	}

	portInt := flag.Int("port", 8080, "TCP port to listen on")
	flag.Parse()

	// if len(os.Args) == 1 {
	// 	flag.Usage()
	// 	return
	// }

	// create the director for uploaded files
	path := dirForPhotos
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err1 := os.Mkdir(path, 0700)
		if err1 != nil {
			log.Fatalf("Create dir '%s' failed:%s", path, err1)
		}
	}

	r := chi.NewRouter()

	registerStaticWeb(r, dirForStatic)

	// r.Get("/", http.FileServer(http.Dir(dirForStatic)).ServeHTTP)

	r.Post("/upload", uploadFileHandler())
	r.Get("/register", mainHandler())
	r.Post("/register", registerHandler())

	port := fmt.Sprintf(":%d", *portInt)
	log.Printf("Server Lepus(v%s) started on port %s", version, port)
	log.Fatal(http.ListenAndServe(port, r))
}
