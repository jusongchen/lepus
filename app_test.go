// Copyright 2019 Jusong Chen

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	signupFormVal := url.Values{"name": {"陈居松"}, "gradYear": {"90"}, "educators": {"%E6%9B%BE%E6%95%8F%E6%AF%85", "余伟然", "%E9%99%88%E7%87%95%E5%8D%8E"}}
	tt := []struct {
		tcName string
		path   string
		method string

		expect   []string
		urlValue url.Values
	}{
		{tcName: "get/", path: "/", method: "GET", expect: []string{"/signup"}},
		{tcName: "get/signup", path: "/signup", method: "GET", expect: []string{`action="/selectphoto"`, `method="post"`}},
		{tcName: "post/selectphoto", path: "/selectphoto", method: "POST", expect: []string{`action="/upload"`, `method="post"`}, urlValue: signupFormVal},
		{tcName: "uploadPhoto", path: "/upload", method: "POST", expect: []string{`action="/where2"`, `method="post"`}},
	}
	baseURL := "http://localhost" + addr
	for _, tc := range tt {
		t.Run(tc.tcName, func(t *testing.T) {
			var err error
			var res *http.Response

			if tc.tcName == "uploadPhoto" {
				// handle file upload case
				extraParams := map[string]string{}

				res, err = postUploadFileRequest(extraParams, "uploadFile", "tests/uploadFile/resources/testPhoto1.jpg", baseURL+tc.path)
				if err != nil {
					t.Fatal(err)
				} else {
					body := &bytes.Buffer{}
					_, err := body.ReadFrom(res.Body)
					if err != nil {
						t.Fatal(err)
					}
					res.Body.Close()
					fmt.Println(res.StatusCode)
					fmt.Println(res.Header)

					fmt.Println(body)
				}
				return
			}

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

			for _, expected := range tc.expect {

				if !strings.Contains(string(bytes.TrimSpace(b)), expected) {
					t.Fatalf("expected the index html; got %s", b)
				}
			}
		})
	}
	app.Stop()

}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func postUploadFileRequest(extraParams map[string]string, nameOfInput string, filePath string, urlPath string) (*http.Response, error) {

	request, err := newfileUploadRequest(urlPath, extraParams, nameOfInput, filePath)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	return client.Do(request)

}
