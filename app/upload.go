package app

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
)

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func resizeImage(src io.Reader, dstFile string) error {

	srcImage, err := imaging.Decode(src)

	if err != nil {
		log.WithError(err).Error("imaging.Decode() failed")
		return err
	}

	dstImage128 := imaging.Resize(srcImage, 256, 0, imaging.Lanczos)

	err = imaging.Save(dstImage128, dstFile)
	if err != nil {
		log.WithError(err).Error("resizeImage():save image failed")
		return err
	}
	log.Infof("resizeImage():save image file to %s", dstFile)
	return nil

}

//uploadFile returns resized image file and error
func (s *lepus) uploadFile(w http.ResponseWriter, r *http.Request) (*UploadReport, error) {

	var err error
	rpt := &UploadReport{}

	defer func() {

		log.WithError(err).WithFields(log.Fields{
			"EndTime":       rpt.EndTime.Format(time.RFC3339),
			"duration":      rpt.Duration,
			"contenSize":    rpt.ContentLength,
			"orginFilename": rpt.OriginName,
			"saveAsFile":    rpt.saveAsName,
			"fileSize":      rpt.FileSize,
		}).Info("upload info")
	}()

	rpt.RealIP = r.RemoteAddr
	rpt.StartTime = time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err = r.ParseMultipartForm(maxUploadSize)
	rpt.EndTime = time.Now()
	rpt.Duration = rpt.EndTime.Sub(rpt.StartTime)
	rpt.ContentLength = r.ContentLength

	if err != nil {

		s.serverErrorWithMsg(w, err, fmt.Sprintf("上传的文件太大（已超过%d兆字节）", maxUploadSize/1024/1024))
		return rpt, err
	}

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return rpt, err
	}
	rpt.OriginName = fileHeader.Filename
	rpt.FileSize = fileHeader.Size

	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("内部错误，无法读取上传文件:%v", err)
		return rpt, err
	}

	fileExtensions := []string{"jpg", "png", "gif", "webp", "cr2", "tif", "bmp", "heif", "jxr", "psd", "ico", "mp4", "m4v", "mkv", "webm", "mov", "avi", "wmv", "mpg", "flv"}

	if filetype.IsImage(fileBytes) {
		rpt.MediaType = imageMedia
	} else if filetype.IsVideo(fileBytes) {
		rpt.MediaType = videoMedia
	} else {
		err = fmt.Errorf("无法识别上传文件的格式,目前支持的文件格式:\n%v\n请将相片或视频转换成支持的格式再上传", strings.Join(fileExtensions, " "))
		return rpt, err
	}

	kind, _ := filetype.Match(fileBytes)
	if kind == filetype.Unknown {
		err = fmt.Errorf("无法识别上传文件的格式")
		return rpt, err
	}

	fileName := randToken(12) + "." + kind.Extension

	// newPath := filepath.Join(s.receiveDir, fileName)

	// // write file
	// newFile, err := os.Create(newPath)
	// if err != nil {
	// 	err = fmt.Errorf("内部错误:create file failed:%v", err)
	// 	return rpt, err
	// }
	// defer newFile.Close() // idempotent, okay to call twice
	// if _, err = newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
	// 	err = fmt.Errorf("内部错误:Write file failed:%v", err)
	// 	return rpt, err
	// }

	rpt.saveAsName = fileName
	rpt.filedata = fileBytes

	return rpt, nil
}
