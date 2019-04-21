// copyright Jusong Chen

package app

import "time"

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

// Media exported
type Media struct {
	MediaType      string
	savedAsFile    string
	OriginFilename string
	Filesize       int
	UploadTime     time.Time
	Data           []byte
}
