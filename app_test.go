// Copyright 2019 Jusong Chen

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"Lepus/app"
)

func TestApp(t *testing.T) {

	addr := ":8888"
	dirForStatic := "./public"
	serverVersion := "v-test-0.1"
	receiveDir := "/tmp/lepus/received"

	// start the app server in a seperate go routine
	go func() {
		app.Start(addr, dirForStatic, serverVersion, receiveDir)
	}()
	val := url.Values{"name": {"陈居松"}, "gradYear": {"90"}, "educators": {"%E6%9B%BE%E6%95%8F%E6%AF%85", "余伟然", "%E9%99%88%E7%87%95%E5%8D%8E"}}
	tt := []struct {
		tcName   string
		path     string
		method   string
		expect   string
		urlValue url.Values
	}{
		{tcName: "home path /", path: "/", method: "GET", expect: "/signup"},
		{tcName: "get /signup", path: "/signup", method: "GET", expect: "/signup"},
		{tcName: "post /signup", path: "/signup", method: "POST", expect: "%E9%99%88%E7%87%95%E5%8D%8E", urlValue: val},
		{tcName: "post /signup", path: "/signup", method: "POST", expect: "余伟然", urlValue: val},
	}
	baseURL := "http://localhost" + addr
	for _, tc := range tt {
		t.Run(tc.tcName, func(t *testing.T) {
			var err error
			var res *http.Response

			if tc.method == "GET" {
				res, err = http.Get(baseURL + tc.path)
			} else if tc.method == "POST" {
				res, err = http.PostForm(baseURL+tc.path, tc.urlValue)
			} else {
				t.Fatalf("unknown http request method %s", tc.method)
			}

			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("expected status OK; got %v", res.Status)
			}

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if !strings.Contains(string(bytes.TrimSpace(b)), tc.expect) {
				t.Fatalf("expected the index html; got %s", b)
			}
		})
	}
	app.Stop()

}
