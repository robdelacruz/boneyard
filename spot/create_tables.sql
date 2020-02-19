PRAGMA foreign_keys = ON;

BEGIN TRANSACTION;
DROP TABLE IF EXISTS task;
CREATE TABLE task (task_id INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, tags TEXT, status INTEGER, createdt TEXT, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES user);

DROP TABLE IF EXISTS note;
CREATE TABLE note (note_id INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, tags TEXT, createdt TEXT, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES user);

DROP TABLE IF EXISTS session;
CREATE TABLE IF NOT EXISTS session (session_key TEXT PRIMARY KEY NOT NULL, user_id INTEGER, FOREIGN KEY(user_id) REFERENCES user);

DROP TABLE IF EXISTS user;
CREATE TABLE user (user_id INTEGER PRIMARY KEY NOT NULL, alias TEXT);

INSERT INTO user (alias) VALUES ('admin');
INSERT INTO user (alias) VALUES ('robdelacruz');

INSERT INTO task (title, body, tags, status, createdt, user_id) VALUES ('task 1', 'Description of task 1', '', 0, '2019-10-13', 1);
END TRANSACTION;

SELECT task_id, title, body, u.user_id, u.alias
FROM task t
LEFT OUTER JOIN user u ON u.user_id = t.user_id;

