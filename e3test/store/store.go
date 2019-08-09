package store

import (
	"database/sql"
	"e3/search"
	"fmt"
	"log"
	"os"

	"github.com/blevesearch/bleve"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	Driver   string
	DSName   string
	IndexDir string
	Logger   *log.Logger
	db       *sql.DB
}

func NewStore(driver, dsname, indexdir string, logger *log.Logger) *Store {
	if logger == nil {
		logger = log.New(os.Stdout, "", 0)
	}

	return &Store{
		Driver:   driver,
		DSName:   dsname,
		IndexDir: indexdir,
		Logger:   logger,
	}
}

func (st *Store) OpenDB() (*sql.DB, error) {
	st.Logger.Printf("Opening driver %s, %s file\n", st.Driver, st.DSName)
	return sql.Open(st.Driver, st.DSName)
}

func (st *Store) DB() *sql.DB {
	if st.db != nil {
		return st.db
	}

	db, err := st.OpenDB()
	if err != nil {
		st.Logger.Printf("error opening db %s, %s (%s)\n", st.Driver, st.DSName, err)
	}
	st.db = db
	return st.db
}

func errSql(q string, err error) error {
	return fmt.Errorf("%s (%s)", err, q)
}

// Execute sql command, with any error occuring added to ErrorBag
func (st *Store) execSql(q string, eb *ErrorBag, vals ...interface{}) {
	s, err := st.DB().Prepare(q)
	if err != nil {
		eb.Add(errSql(q, err))
	}
	if err == nil {
		_, err = s.Exec(vals...)
		if err != nil {
			eb.Add(errSql(q, err))
		}
	}
}

func (st *Store) DropTables() error {
	if st.DB() == nil {
		return dbnilErr()
	}

	var eb ErrorBag
	var q string

	q = "DROP TABLE node"
	st.execSql(q, &eb)
	q = "DROP TABLE nodechange"
	st.execSql(q, &eb)
	q = "DROP TABLE nodetag"
	st.execSql(q, &eb)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) InitTables() error {
	if st.DB() == nil {
		return dbnilErr()
	}

	var eb ErrorBag

	q :=
		`CREATE TABLE IF NOT EXISTS node (
			id TEXT PRIMARY KEY, 
			hash TEXT,
			alias TEXT,
			title TEXT,
			assigned TEXT,
			body TEXT,
			createdt TEXT,
			updatedt TEXT)`
	st.execSql(q, &eb)

	q =
		`CREATE TABLE IF NOT EXISTS nodechange (
			id TEXT PRIMARY KEY)`
	st.execSql(q, &eb)

	q =
		`CREATE TABLE IF NOT EXISTS nodetag (
			id TEXT,
			tag TEXT,
			UNIQUE (id, tag))`
	st.execSql(q, &eb)

	if eb.HasErrors() {
		return eb
	}
	return nil
}

func (st *Store) IndexNode(n *Node) error {
	idx, err := search.BleveIndex(st.IndexDir)
	if err != nil {
		return fmt.Errorf("can't open bleve index (%s)", err)
	}
	defer idx.Close()

	err = idx.Index(n.ID, *n)
	return err
}

func (st *Store) SearchNodes(q string) ([]*Node, error) {
	idx, err := search.BleveIndex(st.IndexDir)
	if err != nil {
		return nil, fmt.Errorf("can't open bleve index (%s)", err)
	}
	defer idx.Close()

	//query := bleve.NewMatchQuery(q)
	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)
	req.Fields = []string{"Title", "Body"}

	results, err := idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("bleve search error (%s)", err)
	}

	var ns []*Node
	for _, match := range results.Hits {
		n, err := st.LoadNodeByID(match.ID)
		if err != nil {
			st.Logger.Printf("load search result node %s error (%s)\n", match.ID, err)
			continue
		}

		if n != nil {
			ns = append(ns, n)
		}
	}

	return ns, nil
}
