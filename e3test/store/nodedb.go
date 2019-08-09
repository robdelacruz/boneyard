package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kjk/betterguid"
	_ "github.com/mattn/go-sqlite3"
)

func isSqlite(st *Store) bool {
	return st.Driver == "sqlite3"
}

func dbnilErr() error {
	return errors.New("error opening db, check logs")
}

func genID() string {
	return betterguid.New()
}

func NewNode() *Node {
	return &Node{}
}

func CloneNode(n *Node) *Node {
	return nil
}

func (st *Store) NodeIsUpToDate(id, hash string) (bool, error) {
	if id == "" {
		return false, nil
	}

	if st.DB() == nil {
		return false, dbnilErr()
	}

	var q string
	if isSqlite(st) {
		q = "SELECT id, hash FROM node where id = ? AND hash = ?"
	} else {
		q = "SELECT id, hash FROM node where id = $1 AND hash = $2"
	}
	row := st.DB().QueryRow(q, id, hash)

	err := row.Scan(&id, &hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("scan sql(%s) error (%s)", q, err)
	}

	return true, nil
}

func (st *Store) LoadNodeByID(id string) (*Node, error) {
	if st.DB() == nil {
		return nil, dbnilErr()
	}

	var q string
	if isSqlite(st) {
		q = "SELECT id, hash, alias, title, assigned, body, createdt, updatedt FROM node WHERE id = ?"
	} else {
		q = "SELECT id, hash, alias, title, assigned, body, createdt, updatedt FROM node WHERE id = $1"
	}
	row := st.DB().QueryRow(q, id)

	n := Node{}
	err := row.Scan(&n.ID, &n.Hash, &n.Alias, &n.Title, &n.Assigned, &n.Body, &n.Createdt, &n.Updatedt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errSql(q, err)
	}

	tags, err := st.LoadNodeTags(n.ID)
	if err != nil {
		return nil, err
	}
	n.Tags = tags

	return &n, nil
}

func (st *Store) LoadNodes(qwhere, qorderby, qlimit string, vals ...interface{}) ([]*Node, error) {
	if st.DB() == nil {
		return nil, dbnilErr()
	}

	q := fmt.Sprintf("SELECT id, hash, alias, title, assigned, body, createdt, updatedt FROM node WHERE %s ORDER BY %s %s", qwhere, qorderby, qlimit)

	rows, err := st.DB().Query(q, vals...)
	if err != nil {
		return nil, errSql(q, err)
	}
	defer rows.Close()

	var ns []*Node
	for rows.Next() {
		var n Node
		err := rows.Scan(&n.ID, &n.Hash, &n.Alias, &n.Title, &n.Assigned, &n.Body, &n.Createdt, &n.Updatedt)
		if err != nil {
			return nil, errSql(q, err)
		}

		tags, err := st.LoadNodeTags(n.ID)
		if err != nil {
			return nil, errSql(q, err)
		}
		n.Tags = tags

		ns = append(ns, &n)
	}

	return ns, nil
}

func (st *Store) insertNode(n *Node) (*Node, error) {
	var q string
	var eb ErrorBag

	if n.ID == "" {
		n.ID = genID()
	}

	nowIsoStr := isotimestr(time.Now())

	if isSqlite(st) {
		q = "INSERT INTO node (id, hash, alias, title, assigned, body, createdt, updatedt) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	} else {
		q = "INSERT INTO node (id, hash, alias, title, assigned, body, createdt, updatedt) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	}
	st.execSql(q, &eb, n.ID, n.HashString(), n.Alias, n.Title, n.Assigned, n.Body, nowIsoStr, nowIsoStr)

	if eb.HasErrors() {
		return nil, eb
	}
	return n, nil
}

func (st *Store) updateNode(n *Node) (*Node, error) {
	var q string
	var eb ErrorBag

	nowIsoStr := isotimestr(time.Now())

	if isSqlite(st) {
		q = "UPDATE node SET hash = ?, alias = ?, title = ?, assigned = ?, body = ?, updatedt = ? WHERE id = ?"
	} else {
		q = "UPDATE node SET hash = $1, alias = $2, title = $3, assigned = $4, body = $5,  updatedt = $6 WHERE id = $7"
	}
	st.execSql(q, &eb, n.HashString(), n.Alias, n.Title, n.Assigned, n.Body, nowIsoStr, n.ID)

	if eb.HasErrors() {
		return nil, eb
	}
	return n, nil
}

func (st *Store) SaveNode(n *Node) (*Node, error) {
	if st.DB() == nil {
		return nil, dbnilErr()
	}

	var err error

	// Insert new node if blank or nonexisting ID, otherwise Update
	if n.ID == "" {
		n, err = st.insertNode(n)
	} else {
		exists, err := st.ExistsNodeID(n.ID)
		if err != nil {
			return nil, err
		}

		if !exists {
			n, err = st.insertNode(n)
		} else {
			n, err = st.updateNode(n)
		}
	}
	if err != nil {
		return nil, err
	}

	// Mark node as changed.
	// Helper processes can use this to find which nodes were
	// updated and rebuild search indexes, related nodes, etc.
	err = st.MarkNodeChanged(n.ID)
	if err != nil {
		return n, err
	}

	// Clear tags first, then add tags one by one,
	//   to simulate an update tags operation.
	// $$ A better way to do this?
	err = st.DeleteNodeTagAll(n.ID)
	if err != nil {
		return nil, err
	}
	for _, tag := range n.Tags {
		err = st.SaveNodeTag(n.ID, tag)
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (st *Store) ExistsTableRow(table, col, val string) (bool, error) {
	var q string
	if isSqlite(st) {
		q = fmt.Sprintf("SELECT %s FROM %s where %s = ?", col, table, col)
	} else {
		q = fmt.Sprintf("SELECT %s FROM %s where %s = $1", col, table, col)
	}
	row := st.DB().QueryRow(q, val)

	var destval interface{}
	err := row.Scan(&destval)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errSql(q, err)
	}

	return true, nil
}

func (st *Store) ExistsNodeID(id string) (bool, error) {
	return st.ExistsTableRow("node", "id", id)
}

func (st *Store) ExistsNodechangeID(id string) (bool, error) {
	return st.ExistsTableRow("nodechange", "id", id)
}

func (st *Store) MarkNodeChanged(id string) error {
	var eb ErrorBag
	var q string
	if isSqlite(st) {
		q = "INSERT OR IGNORE INTO nodechange (id) VALUES (?)"
	} else {
		q = "INSERT INTO nodechange (id) VALUES ($1) ON CONFLICT DO NOTHING"
	}
	st.execSql(q, &eb, id)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) ClearNodeChanged(id string) error {
	if st.DB() == nil {
		return dbnilErr()
	}

	var eb ErrorBag
	var q string
	if isSqlite(st) {
		q = "DELETE FROM nodechange WHERE id = ?"
	} else {
		q = "DELETE FROM nodechange WHERE id = $1"
	}
	st.execSql(q, &eb, id)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) ClearAllNodeChanged() error {
	if st.DB() == nil {
		return dbnilErr()
	}

	var eb ErrorBag
	q := "DELETE FROM nodechange"
	st.execSql(q, &eb)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) QueryChangedNodes() ([]string, error) {
	if st.DB() == nil {
		return nil, dbnilErr()
	}

	q := "SELECT id FROM nodechange"
	rows, err := st.DB().Query(q)
	if err != nil {
		return nil, errSql(q, err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, errSql(q, err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (st *Store) LoadNodeTags(id string) ([]string, error) {
	if st.DB() == nil {
		return nil, dbnilErr()
	}

	var q string
	if isSqlite(st) {
		q = "SELECT tag FROM nodetag WHERE id = ? ORDER BY tag"
	} else {
		q = "SELECT tag FROM nodetag WHERE id = $1 ORDER BY tag"
	}
	rows, err := st.DB().Query(q, id)
	if err != nil {
		return nil, errSql(q, err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, errSql(q, err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (st *Store) SaveNodeTag(id, tag string) error {
	exists, err := st.ExistsNodeID(id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	var eb ErrorBag
	var q string
	if isSqlite(st) {
		q = "INSERT OR IGNORE INTO nodetag (id, tag) VALUES (?, ?)"
	} else {
		q = "INSERT INTO nodetag (id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	}
	st.execSql(q, &eb, id, tag)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) DeleteNodeTagAll(id string) error {
	exists, err := st.ExistsNodeID(id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	var eb ErrorBag
	var q string
	if isSqlite(st) {
		q = "DELETE FROM nodetag WHERE id = ?"
	} else {
		q = "DELETE FROM nodetag WHERE id = $1"
	}
	st.execSql(q, &eb, id)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) DeleteNodeTag(id, tag string) error {
	exists, err := st.ExistsNodeID(id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	var eb ErrorBag
	var q string
	if isSqlite(st) {
		q = "DELETE FROM nodetag WHERE id = ? AND tag = ?"
	} else {
		q = "DELETE FROM nodetag WHERE id = $1 AND tag = $2"
	}
	st.execSql(q, &eb, id, tag)

	if eb.HasErrors() {
		return eb
	}
	return nil
}
