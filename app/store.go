package app

import (
	"encoding/json"
	"fmt"
	"time"

	sql "github.com/jmoiron/sqlx"

	log "github.com/sirupsen/logrus"
)

// Store may be exported as app grow
type Store interface {
	SaveSignup(alumnus *Alumnus, selectedEducators []Educator) error
}

type dbStore struct {
	DB *sql.DB
}

func initSqliteStore(db *sql.DB) error {

	s.store = &dbStore{DB: db}

	sqlStmt := `
	create table if not exists t1(id INTEGER PRIMARY KEY, name varchar(200));
	delete from t1;
	insert into t1 (name) values ('陈居松');
		
	
	create table if not exists alumnus (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), selected_educators text, signup_datetime text);
	
	create table if not exists media (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), media_type text, filename text, filesize integer, origin_filename text, upload_datetime text,upload_duration real, real_ip text,filedata blob);
	
	create table if not exists media_educator(
		media_id integer, 
		edu_name text,
		primary key (media_id,edu_name),
		FOREIGN key (media_id) REFERENCES media(id)
	);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialite sqlite DB")
		return err
	}
	return nil
}

// SaveSignup return alumnusID if suceed
func (s *lepus) SaveSignup(prof AlumnusProfile) (int64, error) {

	sqltext := `insert into alumnus(alumnus_name ,alumnus_gradyear, selected_educators,signup_datetime) values (?,?,?,?)`
	stmt, err := s.store.DB.Prepare(sqltext)
	if err != nil {
		log.WithError(err).Errorf("sql prepare failed:%v", sqltext)
		return 0, err
	}

	educatorsJSON, err := json.Marshal(prof.SelectedEducators)
	if err != nil {
		log.WithError(err).Errorf("Json marshal failed:%v", prof.SelectedEducators)
		return 0, err
	}

	result, err := stmt.Exec(prof.Name, prof.GradYear, string(educatorsJSON), time.Now().Format(time.RFC3339))
	if err != nil {
		log.WithError(err).Errorf("sql execution failed:%v", sqltext)
		return 0, err
	}
	alumnusID, err := result.LastInsertId()
	if err != nil {
		log.WithError(err).Errorf("sql execution could not get LastInsertID:%v", sqltext)
		return 0, err
	}

	log.Infof("Saved sigup info. new alumnus ID:%v, signup:%+v", alumnusID, prof)

	return alumnusID, nil
}

// SaveSignup return alumnusID if suceed
func (s *lepus) SaveUpload(u *UploadReport) error {

	tx, err := s.store.DB.Begin()
	if err != nil {
		log.WithError(err).Errorf("DB begin txn fail!")
		return err
	}

	sqltext := `insert into media 
		(alumnus_name ,alumnus_gradyear , media_type , filename , filesize , origin_filename , upload_datetime, upload_duration,real_ip,filedata)       
		values (?,?,?,?,?,?,?,?,?,?)`
	stmt, err := s.store.DB.Prepare(sqltext)
	if err != nil {
		log.WithError(err).Errorf("sql prepare failed:%v", sqltext)
		return err
	}

	result, err := stmt.Exec(u.Name, u.GradYear, u.MediaType, u.saveAsName, u.FileSize, u.OriginName, u.EndTime.Format(time.RFC3339), u.Duration.Seconds(), u.RealIP, u.filedata)
	if err != nil {
		log.WithError(err).Errorf("sql execution failed:%v", sqltext)
		return err
	}

	mediaID, err := result.LastInsertId()
	if err != nil {
		log.WithError(err).Errorf("sql execution could not get LastInsertID:%v", sqltext)
		return err
	}

	// insert the educator if not exists

	sqltext = `insert into media_educator (media_id , edu_name ) values (?,?)`
	stmt, err = s.store.DB.Prepare(sqltext)
	if err != nil {
		log.WithError(err).Errorf("sql prepare failed:%v", sqltext)
		return err
	}
	for _, n := range u.forEducators {
		_, err := stmt.Exec(mediaID, n)
		if err != nil {
			log.WithError(err).Errorf("sql execution failed:%v", sqltext)
			return err
		}
	}

	tx.Commit()
	log.WithFields(log.Fields{
		"mediaID":         mediaID,
		"name":            u.Name,
		"gradyear":        u.GradYear,
		"mediatype":       u.MediaType,
		"filename":        u.saveAsName,
		"size":            u.FileSize,
		"originname":      u.OriginName,
		"upload_duration": u.Duration.Seconds(),
	}).Info("media saved")

	return nil
}

// getUploadedMedia return []Media
func (s *lepus) getUploadedMedia(from time.Time, to time.Time) ([]Media, error) {

	media := []Media{}

	sqltext := `select id,alumnus_name ,alumnus_gradyear , media_type , filename , filesize , origin_filename , upload_datetime, upload_duration,real_ip 
	from media`

	// result, err := stmt.Exec(prof.Name, prof.GradYear, string(educatorsJSON), time.Now().Format(time.RFC3339))

	err := s.store.DB.Select(&media, sqltext)
	if err != nil {
		log.WithError(err).Errorf("sql execution failed:%v", sqltext)
		return nil, err
	}

	for i := range media {

		eduNames := []string{}
		err := s.store.DB.Select(&eduNames, "select edu_name from media_educator where media_id=?", media[i].MediaID)
		if err != nil {
			log.WithError(err).Errorf("sql execution failed:%v", sqltext)
			return nil, err
		}
		media[i].ForEducators = eduNames
		media[i].FileSizeMb = fmt.Sprintf("%.1f", float64(media[i].FileSize)/1024/1024)
		media[i].UploadRate = float64(media[i].FileSize) / 1024 / 1024 / media[i].Duration
	}

	return media, nil
}

// read media data from DB for a specific mediaID
func (s *lepus) getMediaDataByID(mediaID int64) ([]byte, error) {

	// result, err := stmt.Exec(prof.Name, prof.GradYear, string(educatorsJSON), time.Now().Format(time.RFC3339))
	filedata := []byte{}
	sqltext := `select filedata	from media where id=?`
	err := s.store.DB.Get(&filedata, sqltext, mediaID)
	if err != nil {
		log.WithError(err).Errorf("sql execution failed:%v", sqltext)
		return nil, err
	}
	return filedata, nil
}
