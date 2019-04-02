package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		fileType := r.PostFormValue("type")
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)
		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}

		fileName := randToken(12)
		// fileEndings, err := mime.ExtensionsByType(fileType)
		fileEndings := []string{".jpg"}

		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(dirForPhotos, fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func registerHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if err := r.ParseForm(); err != nil {
			renderError(w, "CANNOT_PARSE_FORM", http.StatusBadRequest)
			return
		}

		fmt.Println("path", r.URL.Path)
		fmt.Println(r.Form) // print information on server side.

		educatorNames := r.Form["educators"]

		fmt.Printf("educator Names %s\n", educatorNames)

		v := NewView("bootstrap", "views/uploadFile.html")

		v.Render(w, educatorNames)

		// http.Redirect(w, r, "/uploadFile.html", http.StatusSeeOther)
	})
}

func mainHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		educatorNames := []string{"蔡春耀", "蔡温榄", "蔡小华", "曹世才", "曾敏毅", "沈淑耀", "陈本培", "陈福音", "陈济兴", "陈金治", "陈群青", "陈细英", "陈燕华", "陈由溪", "陈玉珍", "陈长青", "陈志章", "陈宗辉", "傅德卿", "官尚武", "何天祥", "洪大锋", "胡永模", "胡永岳", "江英岩", "蒋永潮", "李粹玉", "练锡康", "林秉松", "林端川", "林芳", "林嘉坚", "林茂英", "林培坚", "林岫英", "林月心", "林占樟", "林昭英", "刘开天", "刘世煌", "刘永宾", "陆佩珰", "罗朝东", "毛玉珍", "毛祖辉", "毛祖瑜", "欧阳兴", "潘宝升", "潘家健", "潘世英", "潘孝平", "钱振恒", "石柏仁", "童美霞", "王福庆", "王光琳", "王丽华", "王美珠", "王其本", "王强", "王如", "王玉芳", "危金炎", "魏友义", "吴大樑", "吴齐练", "吴生基", "肖方尤", "肖光磊", "肖忠波", "许美钗", "严秀凤", "杨诚", "杨虹", "杨孝华", "杨义森", "叶菁", "余春华", "余冬生", "余世棣", "余天雨", "余伟然", "詹玉赐", "张殿", "张桂贞", "张家新", "张孔烺", "张荣治", "张世年", "张芝萱", "赵文婷", "郑春高", "郑国钦", "郑玉珠", "郑宗振", "周治河", "庄可明", "庄瑞发"}
		// boostrap is a template name defined in layout/boostrap.html
		v := NewView("bootstrap", "views/signup.html")
		v.Render(w, educatorNames)
	})
}
