// copyright Jusong Chen

package app

import "time"

const imageMedia = "image"
const videoMedia = "video"

// Alumnus exposed as this may be referred from other pkg as app grow
type Alumnus struct {
	Name     string
	GradYear string
}

// Educator exposed as this may be referred from other pkg as app grow
type Educator struct {
	Name string `db:"name"`
}

// // Media exported
// type Media struct {
// 	MediaType      string
// 	savedAsFile    string
// 	OriginFilename string
// 	Filesize       int
// 	UploadTime     time.Time
// 	Data           []byte
// }

// AlumnusProfile exposed
type AlumnusProfile struct {
	Alumnus
	SelectedEducators []string
}

// UploadReport is a struct to describe upload info
type UploadReport struct {
	AlumnusProfile
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	ContentLength   int64
	MediaType       string //image, video, audio
	OriginName      string
	saveAsName      string
	FileSize        int64
	resizedFilename string
	RealIP          string
	forEducators    []string
	filedata        []byte
}

// Media is a struct to describe uploaded file
type Media struct {
	MediaID         int64   `db:"id"`
	AlumnusName     string  `db:"alumnus_name"`
	AlumnusGradYear string  `db:"alumnus_gradyear"`
	UploadedTime    string  `db:"upload_datetime"`
	SaveAsName      string  `db:"filename"`
	Duration        float64 `db:"upload_duration"`
	MediaType       string  `db:"media_type"`
	OriginName      string  `db:"origin_filename"`
	FileSize        int64   `db:"filesize"`
	RealIP          string  `db:"real_ip"`
	ForEducators    []string
	FileSizeMb      string
	UploadRate      float64 //in MB/second

	//Filedata        []byte			`db:"filedata"`
}
