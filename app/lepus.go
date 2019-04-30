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

// Educators exposed as this may be referred from other pkg as app grow
type Educators struct {
	Name  string
	Major string
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
	filedata        []byte
}

// Media is a struct to describe uploaded file
type Media struct {
	MediaID         int64
	AlumnusName     string
	AlumnusGradYear string
	UploadedTime    time.Time
	Duration        time.Duration
	MediaType       string //image, video, audio
	OriginName      string
	saveAsName      string
	FileSize        int64
	// filedata        []byte
}
