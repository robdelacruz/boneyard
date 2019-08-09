package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Node map[string]interface{}

func (n Node) String() string {
	var b bytes.Buffer
	for k, v := range n {
		b.WriteString(fmt.Sprintf(".%s %s\n", k, StrV(v)))
	}
	return b.String()
}

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./nbdata.db")
	if err != nil {
		panic(err)
	}
	return db
}

func ExecTrans(db *sql.DB, stmts []string) {
	var err error

	_, err = db.Exec(`BEGIN TRANSACTION;`)
	if err != nil {
		panic(err)
	}

	for _, stmt := range stmts {
		_, err = db.Exec(stmt)
		if err != nil {
			panic(err)
		}
	}

	_, err = db.Exec(`COMMIT;`)
	if err != nil {
		panic(err)
	}

}

func InitTable(db *sql.DB, tbl string) {
	ss := []string{}

	// node table
	ss = append(ss, fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
	id TEXT PRIMARY KEY,
	pubdate TEXT,
	moddate TEXT,
	title TEXT,
	body TEXT)`, tbl))

	// fts full text search table
	ss = append(ss, fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS %s_fts USING fts5(_id, _title, _body, tokenize = porter);`, tbl))

	// insert trigger
	ss = append(ss, fmt.Sprintf(
		`CREATE TRIGGER IF NOT EXISTS %s_insert AFTER INSERT ON %[1]s BEGIN
  INSERT INTO %[1]s_fts (_id, _title, _body)
  VALUES (new.id, new.title, new.body);
	END;`, tbl))

	// update trigger
	ss = append(ss, fmt.Sprintf(
		`CREATE TRIGGER IF NOT EXISTS %s_update AFTER UPDATE OF title, body ON %[1]s BEGIN
	UPDATE %[1]s_fts
	SET _title = new.title, _body = new.body
	WHERE _id = old.id;
	END;`, tbl))

	// delete trigger
	ss = append(ss, fmt.Sprintf(
		`CREATE TRIGGER IF NOT EXISTS %s_delete AFTER DELETE ON %[1]s BEGIN
  DELETE FROM %[1]s_fts WHERE _id = old.id;
	END;`, tbl))

	ExecTrans(db, ss)
}

func PurgeTable(db *sql.DB, tbl string) {
	ss := []string{}
	ss = append(ss, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tbl))
	ss = append(ss, fmt.Sprintf(`DROP TABLE IF EXISTS %s_fts`, tbl))
	ExecTrans(db, ss)
}

func AddNodeTableCols(db *sql.DB, tbl string, cols ...string) {
	tblCols := NodeTableCols(db, tbl)

	for _, col := range cols {
		if SlcContains(tblCols, col) {
			continue
		}

		_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD %s", tbl, col))
		if err != nil {
			panic(err)
		}
	}
}

func NodeTableCols(db *sql.DB, tbl string) []string {
	fields := []string{}

	var sql string
	row := db.QueryRow(`SELECT sql FROM sqlite_master WHERE tbl_name = ?`, tbl)
	row.Scan(&sql)

	r := regexp.MustCompile(`(?s)\((.*)\)`)
	matches := r.FindStringSubmatch(sql)
	if matches == nil {
		return fields
	}

	fieldsSql := strings.TrimSpace(matches[1])
	fieldDefns := regexp.MustCompile(`,\s*`).Split(fieldsSql, -1)

	r = regexp.MustCompile(`\s+`)
	for _, fieldDefn := range fieldDefns {
		field := r.Split(strings.TrimSpace(fieldDefn), -1)[0]
		fields = append(fields, field)
	}
	return fields
}

func NodeTables(db *sql.DB) []string {
	tbls := []string{}
	rows, err := db.Query(`SELECT DISTINCT tbl_name FROM sqlite_master WHERE tbl_name NOT LIKE '%_fts%' ORDER BY tbl_name`)
	if err != nil {
		return tbls
	}

	for rows.Next() {
		var tbl string
		err := rows.Scan(&tbl)
		if err != nil {
			return tbls
		}
		tbls = append(tbls, tbl)
	}
	return tbls
}

func ParseNodeText(snode string) Node {
	return ParseNodeReader(bytes.NewBufferString(snode))
}

func ParseNodeReader(f io.Reader) Node {
	rxpField := regexp.MustCompile(`^\.(\S+)\s+(.*)\s*$`)

	// Parse stdin lines into node struct
	node := Node{}
	bodyLines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := rxpField.FindStringSubmatch(line); matches != nil {
			// Field line - set field value
			// .field1 val
			col := matches[1]
			val := matches[2]
			node[col] = dbVal(val)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error(%s)\n", err)
	}

	node["body"] = strings.Join(bodyLines, "\n")
	return node
}

func ReadNodeCLI(switches map[string]string, parms []string) Node {
	node := Node{}

	if switches["pipe"] != "" {
		// Read body and fields from stdin
		node = ParseNodeReader(os.Stdin)
	}

	if len(parms) > 0 && node["body"] == nil {
		// Body text specified as standalone param.
		node["body"] = strings.Join(parms, "\n")
	}

	// Add any specified field definition.
	// Field definitions are switches starting with a dot (.)
	// Ex. -.title "Note Title" -.author "rob"
	// Fields defined from stdin takes precedence over cli fields.
	for k, v := range switches {
		if strings.HasPrefix(k, ".") {
			col := k[1:]
			if node[col] == nil {
				node[col] = dbVal(v)
			}
		} else if k == "id" {
			node["id"] = v
		}
	}

	return node
}

// Return the right type based on the string value.
// Ex:
// "text1"  => "text1" as string
// "123"    => 123 as int
// "123.45" => 123.45 as float64
func dbVal(v string) interface{} {
	// float
	if regexp.MustCompile(`\d*\.\d+`).MatchString(v) {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}

	// int
	if regexp.MustCompile(`\d+`).MatchString(v) {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}

	// string
	return v
}

// Add a new row to nbdata table.
// node variable contains the map of col => val where col corresponds to the
// database column and val is the value to set.
// Ex.
// If node contains:
// {
//   "title": "Note Title,
//   "body": "Here's a new note.",
//   "amt": 1.23,
//   "tags": "tag1,tag2,",
// }
// This will run the sql query below:
// INSERT INTO nbdata (id, title, body, amt, tags) VALUES (?, ?, ?, ?, ?)
// stmt.Exec(<newid>, "Note Title", "Here's a new note", 1,23, "tag1,tag2,")
//
// Returns:
//   generated ID of new node added, err
//
func AddNode(db *sql.DB, tbl string, node Node) (string, error) {
	// Make sure any body text terminates in newline.
	// That way, appending text to the body will always start on the next line.
	if sbody, ok := node["body"].(string); ok && sbody != "" {
		node["body"] = mustNewlineTerminated(sbody)
	}

	node["pubdate"] = time.Now().UTC().Format(time.RFC3339)
	node["moddate"] = node["pubdate"]

	// Add nonexisting node fields to node table schema
	// Runs: ALTER TABLE nbdata ADD COLUMN <new col>
	tblCols := NodeTableCols(db, tbl)
	nodeCols := MapGetKeys(node)
	newCols := SlcDiff(nodeCols, tblCols)
	AddNodeTableCols(db, tbl, newCols...)

	// Don't add 'SET id = ?' clause. It's not needed.
	nodeCols = SlcRemove(nodeCols, "id")

	// Get each value to set
	// nodeCols... => nodeVals...
	nodeVals := MapGetVals(node, nodeCols)

	// Get placeholder '?' for each nodeCol/nodeVal.
	// nodeCols... => nodeVals... => placeholderVals...
	placeholderVals := extractPlaceholders(nodeCols)

	// nothing to set?
	if len(nodeCols) == 0 {
		return "", fmt.Errorf("nothing to set")
	}

	// Generates 'INSERT INTO nbdata (id, col1, col2...) VALUES (?, ?, ?...)
	sql := fmt.Sprintf("INSERT INTO %s (id, %s) VALUES (?, %s)",
		tbl, strings.Join(nodeCols, ", "), strings.Join(placeholderVals, ", "))
	stmt, _ := db.Prepare(sql)

	// Run INSERT with value params
	newID := GenID()
	execVals := make([]interface{}, 0)
	execVals = append(execVals, newID)
	execVals = append(execVals, nodeVals...)
	_, err := stmt.Exec(execVals...)
	if err != nil {
		return "", err
	}
	return newID, nil
}

// Updates an existing row from the nbdata table.
// node variable contains the map of col => val where col corresponds to the
// database column and val is the value to set.
//
// The node should have an 'id' key which specifies the id of node to update.
//
// Use --overwrite switch to purge all columns from the row before writing
// the specified node column values. If no --overwrite switch, then the
// previous values of the unspecified columns (those not include in node map)
// will be retained.
//
// To increment or decrement a numeric col, use "+<n>" or "-<n>".
// Ex. "count": "+=1"   will add 1 to the current val of 'count' col.
//     "count": "-=1"   will subtract 1 from the current val of 'count' col.
//
// To concatenate to a string col, use "||=<str>".
// Ex. "title": "||= Part 2"  will add " Part 2" to the end of 'title' col.
//
// Ex.
// If node contains:
// {
//   "id": "123",
//   "title": "Note Title,
//   "body": "Here's a new note.",
//   "amt": 1.23,
//   "count": "+=1",
//   "tags": "||=newtag,",
// }
// This will run the sql query below:
// UPDATE nbdata SET title = ?, body = ?, amt = ?, count = count + ?,
//   tags = tags || ? WHERE id = ?
// stmt.Exec("Note Title", "Here's a new note", 1,23, 1, "newtag,", "123")
//
// If --overwrite switch is specified, additional SET col clauses for each
// unspecified col will be added to the sql query to purge the column.
// Ex.
// UPDATE nbdata SET title = ?, body = ?, amt = ?, count = count + ?,
//   tags = tags || ?, debit = '', credit = '', customcol = '' WHERE id = ?
// (where 'debit', 'credit', and 'customcol' are existing nbdata columns that
// was not included as keys in node map.)
//
func UpdateNode(db *sql.DB, tbl string, node Node, switches map[string]string) error {
	// Make sure any body text terminates in newline.
	// That way, appending text to the body will always start on the next line.
	if sbody, ok := node["body"].(string); ok && sbody != "" {
		node["body"] = mustNewlineTerminated(sbody)
	}

	node["moddate"] = time.Now().UTC().Format(time.RFC3339)

	// Add nonexisting node fields to node table schema
	tblCols := NodeTableCols(db, tbl)
	nodeCols := MapGetKeys(node)
	newCols := SlcDiff(nodeCols, tblCols)
	AddNodeTableCols(db, tbl, newCols...)

	id := node["id"]

	// Don't add 'SET id = ?' clause. It's not needed.
	nodeCols = SlcRemove(nodeCols, "id")

	// Get each value to set
	// nodeCols... => nodeVals...
	nodeVals := MapGetVals(node, nodeCols)

	// All specified cols to set
	setClauses := []string{}
	for i, col := range nodeCols {
		if sval, ok := nodeVals[i].(string); ok {
			//
			// col append operations
			// Examples:
			// -.body "||=text to add"    ==> "set body = body || ?"
			// -.tags "||=,tag2,tag3"     ==> "set tags = tags || ?"
			// -.amt +=1.23               ==> "set amt = amt + ?"
			// -.count +=1                ==> "set count = count + ?"
			// -.count -=1                ==> "set count = count - ?"
			//
			if strings.HasPrefix(sval, "+=") {
				nodeVals[i] = sval[2:] // strip out '+='
				setClauses = append(setClauses, fmt.Sprintf("%s = %s + ?", col, col))
				continue
			} else if strings.HasPrefix(sval, "-=") {
				nodeVals[i] = sval[2:]
				setClauses = append(setClauses, fmt.Sprintf("%s = %s - ?", col, col))
				continue
			} else if strings.HasPrefix(sval, "||=") {
				nodeVals[i] = sval[3:] // strip out '||='
				setClauses = append(setClauses, fmt.Sprintf("%s = %s || ?", col, col))
				continue
			}
		}

		// col set
		// -.body "text to set"        ==> "set body = ?"
		// -.amt 1.23                  ==> "set amt = ?"
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
	}

	// --overwrite to completely overwrite node.
	// This will clear out all unspecified node cols.
	if switches["overwrite"] != "" {
		unspecifiedCols := SlcDiff(tblCols, nodeCols)
		unspecifiedCols = SlcRemove(unspecifiedCols, "id")
		unspecifiedCols = SlcRemove(unspecifiedCols, "title")
		for _, col := range unspecifiedCols {
			setClauses = append(setClauses, fmt.Sprintf("%s = ''", col))
		}
	}

	// nothing to set?
	if len(setClauses) == 0 {
		return fmt.Errorf("nothing to set")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		tbl, strings.Join(setClauses, ", "))

	stmt, _ := db.Prepare(sql)
	execVals := make([]interface{}, 0)
	execVals = append(execVals, nodeVals...)
	execVals = append(execVals, id)
	_, err := stmt.Exec(execVals...)
	if err != nil {
		return err
	}
	return nil
}

func DelNode(db *sql.DB, tbl string, id string) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tbl)
	stmt, _ := db.Prepare(sql)
	_, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// Run query on nodes table and return list of matches nodes in specified
// order.
//
// Input:
//   tbl          node table to query (empty string uses default table)
//   selectCols   list of node fields to return
//   orderCols    order by fields, in ascending order
//   where        sql where condition
// Return:
//   nodes        list of nodes
//   nodeCols   node fields used (auto-gen if selectCols nil or empty)
//   error
func QueryNodes(db *sql.DB, tbl string, selectCols []string, orderCols []string, where string, q string) ([]Node, []string, error) {
	nn := []Node{}

	// If not specified, return all nodes, body field last.
	if selectCols == nil || len(selectCols) == 0 {
		selectCols = NodeTableCols(db, tbl)

		if SlcContains(selectCols, "body") {
			selectCols = SlcRemove(selectCols, "body")
			selectCols = append(selectCols, "body")
		}
	}
	// If not specified, return nodes in the order they were added.
	if len(orderCols) == 0 {
		orderCols = append(orderCols, "id")
	}
	if strings.TrimSpace(where) == "" {
		where = "1"
	}

	// regex for "SUM(amt) AS amt"
	r := regexp.MustCompile(`(?i)^\s*(?:SUM|AVG|COUNT)\((\w+)\)\s+AS\s+\w+\s*$`)

	// Include only existing node cols
	nodeCols := []string{}
	tblCols := NodeTableCols(db, tbl)
	for _, selectCol := range selectCols {
		// Get field from SUM/AVG select
		// Ex. "SUM(amt) AS amt"  extracts "amt" field
		matches := r.FindStringSubmatch(selectCol)
		if matches != nil {
			col := matches[1]
			if SlcContains(tblCols, col) {
				nodeCols = append(nodeCols, selectCol)
			}
			continue
		}

		if SlcContains(tblCols, selectCol) {
			nodeCols = append(nodeCols, selectCol)
		}
	}

	// If no select cols exist in node, default to id.
	if len(nodeCols) == 0 {
		nodeCols = append(nodeCols, "id")
	}

	// Sql query takes two forms, depending if q parm specified:
	//
	// (a) Full text search with where clause - this will join
	//     nbdata table with fts table to 'match' the full text search
	//     string, and apply a where clause to nbdata table.
	//
	// (b) Just a where clause applied to nbdata table.
	//
	var sql string
	if q != "" {
		// Full text search
		sql = fmt.Sprintf(
			`SELECT %s FROM %s AS d 
			INNER JOIN %[2]s_fts AS fts ON id = fts._id 
			WHERE %[2]s_fts MATCH '%s' AND %s 
			ORDER BY %s`,
			strings.Join(nodeCols, ", "),
			tbl,
			q,
			where,
			strings.Join(orderCols, ", "))
	} else {
		// Just where clause query
		sql = fmt.Sprintf(
			`SELECT %s FROM %s 
			WHERE %s 
			ORDER BY %s`,
			strings.Join(nodeCols, ", "),
			tbl,
			where,
			strings.Join(orderCols, ", "))
	}

	rows, err := db.Query(sql)
	if err != nil {
		return nn, nil, err
	}

	// Hacky way to create slice of pointer to interfaces for rows.Scan()
	vv := make([]interface{}, len(nodeCols))
	pp := make([]interface{}, len(nodeCols))
	for i, _ := range vv {
		pp[i] = &vv[i]
	}

	cols, _ := rows.Columns()
	for rows.Next() {
		// Row values go in pp[] which are pointers to each element of vv[]
		err = rows.Scan(pp...)
		if err != nil {
			return nn, nil, err
		}
		n := Node{}
		for i, v := range vv {
			// convert []byte to string
			if bs, ok := v.([]byte); ok {
				v = string(bs)
			}

			n[cols[i]] = v
		}
		nn = append(nn, n)
	}
	rows.Close()

	return nn, nodeCols, nil
}

//--- helper funcs ---

func extractPlaceholders(kk []string) []string {
	vv := []string{}
	for range kk {
		vv = append(vv, "?")
	}
	return vv
}

// Make sure line is terminated with a newline. Append newline if necessary.
func mustNewlineTerminated(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
