package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Tags         string `json:"tags"`
	Status       int    `json:"status"`
	Createdt     string `json:"createdt"`
	CreateUserID int    `json:"createuserid"`
	AssignUserID int    `json:"assignuserid"`
}

func (t Task) String() string {
	return fmt.Sprintf(`
ID:        %d
Title:     %s
Body:      %s
Tags:      %s
Status:    %d
Createdt:  %s
CreateUserID: %d
AssignUserID: %d
`, t.ID, t.Title, t.Body, t.Tags, t.Status, t.Createdt, t.CreateUserID, t.AssignUserID)
}

func NewTask() Task {
	t := Task{}
	t.Createdt = time.Now().Format(time.RFC3339)
	t.Title = fmt.Sprintf("Task %s", t.Createdt)
	return t
}

func openDB(switches map[string]string) (*sql.DB, error) {
	// spot db file is read from the following, in order of priority:
	// 1. -F <file>
	// 2. SPOTDB env var
	// 3. /usr/local/share/spot/spot.db (default)
	dbfile := os.Getenv("SPOTDB")
	if switches["F"] != "" {
		dbfile = switches["F"]
	}
	if dbfile == "" {
		dirpath := filepath.Join(string(os.PathSeparator), "usr", "local", "share", "spot")
		os.MkdirAll(dirpath, os.ModePerm)
		dbfile = filepath.Join(dirpath, "spot.db")
	}

	fmt.Printf("dbfile: '%s'\n", dbfile)
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, fmt.Errorf("openDB(): Error opening '%s' (%s)\n", dbfile, err)
	}

	ensureCreateTables(db)
	return db, nil
}

func ensureCreateTables(db *sql.DB) {
	sqlstr := `PRAGMA foreign_keys = ON;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS task (taskid INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, tags TEXT, status INTEGER, createdt TEXT, create_userid INTEGER, assign_userid INTEGER);
CREATE TABLE IF NOT EXISTS note (noteid INTEGER PRIMARY KEY NOT NULL, title TEXT, body TEXT, tags TEXT, createdt TEXT, create_userid INTEGER);
CREATE TABLE IF NOT EXISTS user (userid INTEGER PRIMARY KEY NOT NULL, alias TEXT);
CREATE TABLE IF NOT EXISTS session (sessionkey TEXT PRIMARY KEY NOT NULL, userid INTEGER);
END TRANSACTION;`
	_, err := db.Exec(sqlstr)
	if err != nil {
		panic(err)
	}
}

func insertTask(db *sql.DB, t *Task) error {
	sqlstr := `INSERT INTO task (title, body, tags, status, createdt, create_userid, assign_userid) VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmt, _ := db.Prepare(sqlstr)
	result, err := stmt.Exec(t.Title, t.Body, t.Tags, t.Status, t.Createdt, t.CreateUserID, t.AssignUserID)
	if err != nil {
		return fmt.Errorf("insertTask() -\n%v\n(%s)\n", t, err)
	}
	t.ID, _ = result.LastInsertId()
	return nil
}

func updateTask(db *sql.DB, t *Task) error {
	sqlstr := `UPDATE task set title = ?, body = ?, tags = ?, status = ?, assign_userid = ? WHERE taskid = ?`
	stmt, _ := db.Prepare(sqlstr)
	_, err := stmt.Exec(t.Title, t.Body, t.Tags, t.Status, t.AssignUserID, t.ID)
	if err != nil {
		return fmt.Errorf("updateTask() -\n%v\n(%s)\n", t, err)
	}
	return nil
}

func selectTask(db *sql.DB, qwhere string) ([]*Task, error) {
	if qwhere == "" {
		qwhere = "1=1"
	}
	qorderby := "taskid"
	sqlstr := fmt.Sprintf(`SELECT taskid, title, body, tags, status, createdt, create_userid, assign_userid FROM task WHERE %s ORDER BY %s`, qwhere, qorderby)

	rows, err := db.Query(sqlstr)
	if err != nil {
		return nil, fmt.Errorf("selectTask() - (%s)\n", err)
	}
	tt := []*Task{}
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Title, &t.Body, &t.Tags, &t.Status, &t.Createdt, &t.CreateUserID, &t.AssignUserID)
		tt = append(tt, &t)
	}

	return tt, nil
}

func selectOneTask(db *sql.DB, taskID int64) (*Task, error) {
	sqlstr := `SELECT taskid, title, body, tags, status, createdt, create_userid, assign_userid FROM task WHERE taskid = ?`
	row := db.QueryRow(sqlstr, taskID)

	var t Task
	err := row.Scan(&t.ID, &t.Title, &t.Body, &t.Tags, &t.Status, &t.Createdt, &t.CreateUserID, &t.AssignUserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("selectOneTask() - (%s)\n", err)
	}
	return &t, nil
}
