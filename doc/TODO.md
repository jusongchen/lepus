
* search uploaded
* image rotated
* DB support
* tls?



create table t1(id INTEGER PRIMARY KEY, name varchar(200));
insert into t1 values ('陈居松');


create TABLE educator(id INTEGER PRIMARY KEY ,name text,subject text);

create table alumnus (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), signup_datetime text);

create table alumnus_educator (
alumnus_id integer, 
educator_id integer,
primary key (alumnus_id,educator_id),
FOREIGN key (alumnus_id) REFERENCES alumnus(id),
FOREIGN key (educator_id) REFERENCES educator(id)
);

create table media (id INTEGER PRIMARY KEY,alumnus_name text,alumnus_gradyear char(2), media_type text, filename text, filesize integer, origin_filename text, upload_datetime text, filedata blob);


create table media_educator(
media_id integer, 
educator_id integer,
primary key (media_id,educator_id),
FOREIGN key (media_id) REFERENCES media(id),
FOREIGN key (educator_id) REFERENCES educator(id)
);
