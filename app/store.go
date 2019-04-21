package app

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

// Store may be exported as app grow
type Store interface {
	SaveSignup(alumnus *Alumnus, selectedEducators []Educators) error
}

type dbStore struct {
	DB *sql.DB
}

// // The store variable is a package level variable that will be available for
// // use throughout our application code
// var store Store

// // InitStore method is to initialize the store. when the server starts up
// func InitStore(s Store) {
// 	store = s
// }

// not found
func initSqliteStore(db *sql.DB) error {

	s.store = &dbStore{DB: db}

	sqlStmt := `
	create table if not exists t1(id INTEGER PRIMARY KEY, name varchar(200));
	delete from t1;
	insert into t1 (name) values ('陈居松');
		
	create TABLE if not exists educator(id INTEGER PRIMARY KEY ,name text,subject text);
	
	create table if not exists alumnus (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), signup_datetime text);
	
	create table if not exists alumnus_educator (
	alumnus_id integer, 
	educator_id integer,
	primary key (alumnus_id,educator_id),
	FOREIGN key (alumnus_id) REFERENCES alumnus(id),
	FOREIGN key (educator_id) REFERENCES educator(id)
	);
	
	create table if not exists media (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), media_type text, filename text, filesize integer, origin_filename text, upload_datetime text, filedata blob);
	
	
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
		logrus.WithError(err).Fatal("Failed to initialite sqlite DB")
		return err
	}
	return nil
}
