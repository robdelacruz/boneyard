PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS notereply;
DROP TABLE IF EXISTS note;
DROP TABLE IF EXISTS wgroupuser;
DROP TABLE IF EXISTS wgroup;
DROP TABLE IF EXISTS user;

CREATE TABLE IF NOT EXISTS wgroup (
    groupname TEXT NOT NULL PRIMARY KEY,
    fullname TEXT
);

CREATE TABLE IF NOT EXISTS user (
    username TEXT NOT NULL PRIMARY KEY,
    fullname TEXT
);

CREATE TABLE IF NOT EXISTS wgroupuser (
    groupname TEXT NOT NULL,
    username TEXT NOT NULL,
    FOREIGN KEY(groupname) REFERENCES wgroup,
    FOREIGN KEY(username) REFERENCES user,
    PRIMARY KEY(groupname, username)
);

CREATE TABLE IF NOT EXISTS note (
    noteid INTEGER NOT NULL PRIMARY KEY,
    groupname TEXT NOT NULL,
    username TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    FOREIGN KEY(groupname) REFERENCES wgroup,
    FOREIGN KEY(username) REFERENCES user
);

CREATE TABLE IF NOT EXISTS notereply (
    notereplyid INTEGER NOT NULL PRIMARY KEY,
    noteid INTEGER NOT NULL,
    username TEXT NOT NULL,
    body TEXT NOT NULL,
    FOREIGN KEY(noteid) REFERENCES note,
    FOREIGN KEY(username) REFERENCES user
);

.mode csv

.import wgroup.csv wgroup
.import user.csv user
.import wgroupuser.csv wgroupuser
.import note.csv note
.import notereply.csv notereply
