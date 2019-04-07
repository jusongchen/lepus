// copyright Jusong Chen

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"Lepus/app"
)

const version = "0.3"
const defaultPort = "8080"
const receiveDir = "./received"
const dirForStatic = "./public"

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
	if _, err := os.Stat(receiveDir); os.IsNotExist(err) {
		err1 := os.Mkdir(receiveDir, 0700)
		if err1 != nil {
			log.Fatalf("Create dir '%s' failed:%s", receiveDir, err1)
		}
	}

	port := fmt.Sprintf(":%d", *portInt)
	app.Start(port, dirForStatic, version, receiveDir)

}
