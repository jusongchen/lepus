package app

import (
	"database/sql"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

// Store may be exported as app grow
type Store interface {
	SaveSignup(alumnus *Alumnus, selectedEducators []Educators) error
}

type dbStore struct {
	DB *sql.DB
}

// create table if not exists alumnus_educator (
// alumnus_id integer,
// educator_id integer,
// primary key (alumnus_id,educator_id),
// FOREIGN key (alumnus_id) REFERENCES alumnus(id),
// FOREIGN key (educator_id) REFERENCES educator(id)
// );

func initSqliteStore(db *sql.DB) error {

	s.store = &dbStore{DB: db}

	sqlStmt := `
	create table if not exists t1(id INTEGER PRIMARY KEY, name varchar(200));
	delete from t1;
	insert into t1 (name) values ('陈居松');
		
	create TABLE if not exists educator(id INTEGER PRIMARY KEY ,name text,subject text);
	
	create table if not exists alumnus (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), selected_educators text, signup_datetime text);
	
	create table if not exists media (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), media_type text, filename text, filesize integer, origin_filename text, upload_datetime text,upload_duration real, real_ip text,filedata blob);
	
	create table if not exists media_educator(
	media_id integer, 
	educator_id integer,
	primary key (media_id,educator_id),
	FOREIGN key (media_id) REFERENCES media(id),
	FOREIGN key (educator_id) REFERENCES educator(id)
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

	sqltext = `insert into media_educator (media_id , educator_id ) values (?,?)`
	stmt, err = s.store.DB.Prepare(sqltext)
	if err != nil {
		log.WithError(err).Errorf("sql prepare failed:%v", sqltext)
		return err
	}

	for _, name := range u.SelectedEducators {
		_, err := stmt.Exec(mediaID, name)
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
