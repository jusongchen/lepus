package app

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
)

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func resizeImage(srcFile, dstFile string) error {

	srcImage, err := imaging.Open(srcFile)

	if err != nil {
		logrus.WithError(err).Error("resizeImage():open image failed")
		return err
	}

	dstImage128 := imaging.Resize(srcImage, 128, 128, imaging.Lanczos)

	err = imaging.Save(dstImage128, dstFile)
	if err != nil {
		logrus.WithError(err).Error("resizeImage():save image failed")
		return err
	}
	logrus.Infof("resizeImage():save image file to %s", dstFile)
	return nil

}

// UploadReport is a struct to describe upload info
type UploadReport struct {
	startTime         time.Time
	endTime           time.Time
	duration          time.Duration
	httpContentLength int64
	originName        string
	saveAsName        string
	fileSize          int64
	resizedFilename   string
}

//uploadFile returns resized image file and error
func (s *lepus) uploadFile(w http.ResponseWriter, r *http.Request) (*UploadReport, error) {

	var err error
	rpt := &UploadReport{}

	rpt.startTime = time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err = r.ParseMultipartForm(maxUploadSize)
	rpt.endTime = time.Now()
	rpt.duration = rpt.endTime.Sub(rpt.startTime)
	rpt.httpContentLength = r.ContentLength

	logrus.WithError(err).WithFields(logrus.Fields{
		"endTime":            rpt.endTime.Format(time.RFC3339),
		"duration":           rpt.duration,
		"size":               r.ContentLength,
		"rate(bytes/second)": float64(r.ContentLength) / rpt.duration.Seconds(),
	}).Info("file upload completed")

	if err != nil {
		// renderError(w, msg,: http.StatusBadRequest)
		err = fmt.Errorf("上传的文件太大（已超过%d兆字节）:%v", maxUploadSize/1024/1024, err)
		return rpt, err
	}

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return rpt, err
	}

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return rpt, err
	}

	fileExtensions := []string{"jpg", "png", "gif", "webp", "cr2", "tif", "bmp", "heif", "jxr", "psd", "ico", "mp4", "m4v", "mkv", "webm", "mov", "avi", "wmv", "mpg", "flv"}

	isImage := filetype.IsImage(fileBytes)
	if !(isImage || filetype.IsVideo(fileBytes) || filetype.IsAudio(fileBytes)) {
		err = fmt.Errorf("无法识别上传文件的格式,目前支持的文件格式:\n%v\n请将相片或视频转换成支持的格式再上传", strings.Join(fileExtensions, " "))
		return rpt, err
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == filetype.Unknown {
		err = fmt.Errorf("无法识别上传文件的格式")
		return rpt, err
	}

	fileName := randToken(12) + "." + kind.Extension

	newPath := filepath.Join(s.receiveDir, fileName)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		err = fmt.Errorf("内部错误:create file failed:%v", err)
		return rpt, err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		err = fmt.Errorf("内部错误:Write file failed:%v", err)
		return rpt, err
	}

	rpt.originName = fileHeader.Filename
	rpt.saveAsName = fileName
	rpt.fileSize = fileHeader.Size

	if !isImage {
		return rpt, err
	}

	if err = resizeImage(newPath, filepath.Join(s.staticHomeDir, s.imageDir, fileName)); err != nil {
		// just log error, we may get an error during resize the picture as we do not handle all formats
		logrus.WithError(err).WithField("filename", newPath).Error("resize image failed")
		//do not return error here, as even resize failed, we still move forward
		err = nil
		return rpt, err
	}
	rpt.resizedFilename = fileName
	return rpt, nil
}
