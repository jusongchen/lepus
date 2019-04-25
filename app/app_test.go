// Copyright 2019 Jusong Chen

package app_test

import (
	"bytes"
	"database/sql"
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

	"github.com/jusongchen/lepus/app"
	"github.com/sirupsen/logrus"
)

func TestApp(t *testing.T) {

	addr := ":8888"
	dirForStatic := "./public"
	imageDir := "./images"
	receiveDir := "/tmp/lepus/received"
	const viewPath = "../views"
	const sqliteFile = "lepus.test.sqlite.db"

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatalf("Faile top open sqlite3 DB:%v", err)
	}

	sessionKey := "Tes1ses2ion3eys4uey5IWU6hdu7j&%8"

	lepusHomeDir, _ := os.Getwd()
	// start the app server in a seperate go routine
	go func() {
		var educatorNames = []string{"蔡春耀", "蔡温榄", "蔡小华", "曹世才", "曾敏毅", "沈淑耀", "陈本培", "陈福音", "陈济兴", "陈金治", "陈群青", "陈细英", "陈燕华", "陈由溪", "陈玉珍", "陈长青", "陈志章", "陈宗辉", "傅德卿", "官尚武", "何天祥", "洪大锋", "胡永模", "胡永岳", "江英岩", "蒋永潮", "李粹玉", "练锡康", "林秉松", "林端川", "林芳", "林嘉坚", "林茂英", "林培坚", "林岫英", "林月心", "林占樟", "林昭英", "刘开天", "刘世煌", "刘永宾", "陆佩珰", "罗朝东", "毛玉珍", "毛祖辉", "毛祖瑜", "欧阳兴", "潘宝升", "潘家健", "潘世英", "潘孝平", "钱振恒", "石柏仁", "童美霞", "王福庆", "王光琳", "王丽华", "王美珠", "王其本", "王强", "王如", "王玉芳", "危金炎", "魏友义", "吴大樑", "吴齐练", "吴生基", "肖方尤", "肖光磊", "肖忠波", "许美钗", "严秀凤", "杨诚", "杨虹", "杨孝华", "杨义森", "叶菁", "余春华", "余冬生", "余世棣", "余天雨", "余伟然", "詹玉赐", "张殿", "张桂贞", "张家新", "张孔烺", "张荣治", "张世年", "张芝萱", "赵文婷", "郑春高", "郑国钦", "郑玉珠", "郑宗振", "周治河", "庄可明", "庄瑞发"}
		app.Start(db, sessionKey, addr, lepusHomeDir, dirForStatic, receiveDir, imageDir, viewPath, educatorNames)
	}()
	signupFormVal := url.Values{"name": {"陈居松"}, "gradYear": {"90"}, "educators": {"%E6%9B%BE%E6%95%8F%E6%AF%85", "余伟然", "%E9%99%88%E7%87%95%E5%8D%8E"}}
	tt := []struct {
		tcName string
		path   string
		method string

		expect   []string
		urlValue url.Values
	}{
		// {tcName: "get/", path: "/", method: "GET", expect: []string{`action="/signup"`}},
		{tcName: "get/signup", path: "/signup", method: "GET", expect: []string{`action="/selectphoto"`, `method="post"`}},
		{tcName: "post/selectphoto", path: "/selectphoto", method: "POST", expect: []string{`action="/upload"`, `method="post"`, `sessionID={"`}, urlValue: signupFormVal},
		{tcName: "uploadPhoto", path: "/where2", method: "POST", expect: []string{`action="/where2"`, `method="post"`, `sessionID={"`}},
	}
	baseURL := "http://localhost" + addr
	for _, tc := range tt {
		t.Run(tc.tcName, func(t *testing.T) {
			var err error
			var res *http.Response

			if tc.tcName == "uploadPhoto" {
				// handle file upload case
				extraParams := map[string]string{
					"sessionID": "{&quot;Name&quot;:&quot;JUS&quot;,&quot;GradYear&quot;:&quot;65&quot;,&quot;SelectedEducators&quot;:[&quot;蔡春耀&quot;,&quot;蔡温榄&quot;,&quot;陈本培&quot;,&quot;陈金治&quot;]}",
				}

				res, err = postUploadFileRequest(extraParams, "uploadFile", "tests/resources/testPhoto1.jpg", baseURL+tc.path)
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
		logrus.Fatal(err)
	}
	client := &http.Client{}
	return client.Do(request)

}
