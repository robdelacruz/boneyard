package main

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func errFail(t *testing.T, err error) {
	if err != nil {
		t.Errorf("error (%s)", err)
	}
}

var db *sql.DB
var loremipsum string
var nodeMismatchFmt string

func init() {
	loremipsum =
		`Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
	
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`

	nodeMismatchFmt = `add / find mismatch:
Expected:
%s
Actual:
%s`

	db = InitDB()
	PurgeTable(db, "t1")
}

func TestAddEdit(t *testing.T) {
	fmt.Println("TestAddEdit...")

	// Add node
	s := fmt.Sprintf(`add -tbl t1 -.title "Node Title 123@#$-+ 123" -.amt 123.45 -.body "%s" -.tags "tag1,tag2,multi-word tag,tag3"`, loremipsum)
	res, err := RunStmt(db, s)
	errFail(t, err)
	if res.NodeID == "" {
		t.Fatalf("[add] no node id returned")
	}
	newID := res.NodeID

	// Verify that node was added
	s = `find --silent -tbl t1 -select id,title,amt,body`
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	nActual := res.Nodes[0]
	nExpected := ParseNodeText(fmt.Sprintf(
		`.id %s
.title Node Title 123@#$-+ 123
.amt 123.45
.tags tag1,tag2,multi-word tag,tag3
%s

`, newID, loremipsum))

	if !MapContains(nExpected, nActual) {
		t.Errorf(nodeMismatchFmt, nExpected, nActual)
	}

	// '\n' should be auto-appended to body if body didn't end with newline
	if !strings.HasSuffix(nActual["body"].(string), "\n") {
		t.Errorf("body should end with newline (newline auto-added)")
	}

	// Edit node
	s = fmt.Sprintf(`edit -id %s -tbl t1 -.title "Node Title Edited" -.amt 99`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)

	// Verify that node was edited
	s = fmt.Sprintf(`find --silent -tbl t1 -sel title,amt -where "id = '%s'"`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	n := res.Nodes[0]
	if n["title"] != "Node Title Edited" {
		t.Errorf("expected '%s', actual: '%s'", "Node Title Edited", n["title"])
	}
	if n["amt"].(int64) != 99 {
		t.Errorf("expected %d, actual: %f", 99, n["amt"])
	}
}

func TestDel(t *testing.T) {
	fmt.Println("TestDel...")

	// Add something to delete later
	s := `add -tbl t1 -.title "Delete me" -.amt 123.45`
	res, err := RunStmt(db, s)
	errFail(t, err)

	if res.NodeID == "" {
		t.Fatalf("No new id returned")
	}
	nodeID := res.NodeID

	// Verify that node was added
	s = `find --silent -tbl t1 -select id,title -w "title = 'Delete me'"`
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	readNode := res.Nodes[0]

	if readNode["id"] != nodeID {
		t.Errorf("Read node id not the same as ID returned")
	}

	// Delete the node
	s = fmt.Sprintf(`del -tbl t1 -id %s`, nodeID)
	_, err = RunStmt(db, s)
	errFail(t, err)

	// Verify that node was deleted
	s = `find --silent -tbl t1 -select id,title -where "title = 'Delete me'"`
	res, err = RunStmt(db, s)
	errFail(t, err)

	if len(res.Nodes) != 0 {
		t.Errorf("Node wasn't deleted.")
	}
}

func TestFindTotals(t *testing.T) {
	fmt.Println("TestFindTotals...")

	//
	// Add expense nodes to be queried and subtotaled later.
	//

	s := `add -tbl t1 -.title "jeep" -.cat "commute" -.amt 9.0 -.tags "tag1,test123,"`
	_, err := RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "mcdo" -.cat "dine_out" -.amt 255.50 -.tags "test123,tag2,"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "wendy's" -.cat "dine_out" -.amt 199.25 -.tags "wendy's,fast food,test123,"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "rustan's" -.cat "grocery" -.amt 10000.25 -.tags "rustan's,test123,"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "NBA League Pass" -.cat "household" -.amt 4999 -.tags "test123,nba"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "mostly cat food" -.cat "grocery" -.amt 700 -.tags "cat food,pets,test123,cats,"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.title "Burger King" -.cat "dine_out" -.amt 125.50 -.tags "test123,fast food,"`
	_, err = RunStmt(db, s)
	errFail(t, err)

	// Total all cat:dine_out
	s = `find --silent -tbl t1 -select title,cat,amt,tags -w "cat = 'dine_out' and tags like '%test123%'" -total amt -o amt,title`
	res, err := RunStmt(db, s)
	errFail(t, err)

	if len(res.Nodes) != 3 {
		t.Errorf("Total all cat = 'dine_out'\nexpected: 3 nodes, actual: %d node", len(res.Nodes))
	}

	if res.SumNode == nil {
		t.Fatalf("No sum node for: total all cat = 'dine_out'")
	}

	sumAmt, ok := res.SumNode["amt"].(float64)
	if !ok {
		t.Errorf("Sum amt is not float type")
	}
	expectedSumAmt := 255.50 + 199.25 + 125.50
	if sumAmt != expectedSumAmt {
		t.Errorf("Sum amt incorrect. expected: %f, actual: %f", expectedSumAmt, sumAmt)
	}

	// Total all test123 order by title,cat
	s = `find --silent -tbl t1 -s title,cat,amt -where "tags like '%test123%'" -order title,cat -t amt`
	res, err = RunStmt(db, s)
	errFail(t, err)

	if len(res.Nodes) != 7 {
		t.Fatalf("Total all cat = 'dine_out'\nexpected: 7 nodes, actual: %d node", len(res.Nodes))
	}

	if res.SumNode == nil {
		t.Fatalf("No sum node for: total all test123 order by title,cat")
	}

	sumAmt, ok = res.SumNode["amt"].(float64)
	if !ok {
		t.Errorf("Sum amt is not float type")
	}
	expectedSumAmt = 9.0 + 255.50 + 199.25 + 10000.25 + 4999 + 700 + 125.50
	if sumAmt != expectedSumAmt {
		t.Errorf("Sum amt incorrect. expected: %f, actual: %f", expectedSumAmt, sumAmt)
	}

	firstNode := res.Nodes[0]
	lastNode := res.Nodes[len(res.Nodes)-1]

	if firstNode["title"] != "Burger King" {
		t.Errorf("First node sorted by title,cat incorrect.")
	}
	if lastNode["title"] != "wendy's" {
		t.Errorf("Last node sorted by title,cat incorrect.")
	}
}

func TestMissingCols(t *testing.T) {
	fmt.Println("TestMissingCols...")

	PurgeTable(db, "t1")

	s := `add -tbl t1 -.newcol1 val1 -.newcol2 val2`
	_, err := RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.newcol1 val3 -.newcol2 val4`
	_, err = RunStmt(db, s)
	errFail(t, err)

	s = `add -tbl t1 -.newcol1 val3 -.newcol2 val4 -.newcol3 val5`
	_, err = RunStmt(db, s)
	errFail(t, err)

	// Find with existing cols should not return error
	s = `find --silent -tbl t1 -select id,newcol1,newcol2,newcol3 -o id`
	_, err = RunStmt(db, s)
	errFail(t, err)

	// Find with missing col 'newcolx' should ignore newcolx and newcolx field
	// should not be present in returned nodes.
	s = `find --silent -tbl t1 -sel id,newcol1,newcol2,newcolx,newcol3`
	res, err := RunStmt(db, s)
	errFail(t, err)

	if len(res.Nodes) == 0 {
		t.Errorf("should return nodes")
	}
	for _, n := range res.Nodes {
		if _, ok := n["newcolx"]; ok {
			t.Errorf("missing col 'newcolx' should not be in returned node")
			break
		}
		if _, ok := n["newcol3"]; !ok {
			t.Errorf("existing col 'newcol3' should be present in returned node")
			break
		}
	}
}

func TestAppend(t *testing.T) {
	fmt.Println("TestAppend...")

	// Add node
	s := `add -tbl t1 -.title "append test" -.amt 100.0 -.body "line1" -.tags "tag1,"`
	res, err := RunStmt(db, s)
	errFail(t, err)
	if res.NodeID == "" {
		t.Fatalf("[add] no node id returned")
	}
	newID := res.NodeID

	// Edit node, append to body, amt and unspecified fields fieldx and fieldy
	// Should not return error even when appending to unspecified fields,
	s = fmt.Sprintf(`edit -id %s -tbl t1 -.body "||=line2" -.amt +=9.99 -.fieldx "||=abc" -.fieldy +=1.23`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)

	// Read updated node
	s = fmt.Sprintf(`find --silent -tbl t1 -s body,amt,tags -where "id = '%s'"`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	n := res.Nodes[0]

	// Verify that body and amt were appended to.
	// Note that body should be auto-terminated with newline always.
	if n["body"] != "line1\nline2\n" {
		t.Errorf("expected body =\n'%s', actual body =\n'%s'", "line1\nline2\n", n["body"])
	}
	if n["amt"] != 109.99 {
		t.Errorf("expected amt = %f, actual amt = %f", 109.99, n["amt"])
	}

	// tags should be unchanged
	if n["tags"] != "tag1," {
		t.Errorf("tags should be unchanged since it wasn't appended to")
	}

	// Edit node, append to tags
	s = fmt.Sprintf(`edit -id %s -tbl t1 -.tags "||=tag2,"`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)

	// Read updated node
	s = fmt.Sprintf(`find --silent -tbl t1 -s tags -w "id = '%s'"`, newID)
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	n = res.Nodes[0]

	// Verify that tags were appended to.
	if n["tags"] != "tag1,tag2," {
		t.Errorf("expected tags = '%s', actual tags = '%s'", "tag1,tag2,", n["tags"])
	}
}

func TestFindNonexistingTable(t *testing.T) {
	fmt.Println("TestFindNonexistingTable...")

	PurgeTable(db, "tblx")
	PurgeTable(db, "tbly")
	PurgeTable(db, "tblz")

	// Add node on nonexistent table should create the table
	s := `add -tbl tblx -.title "Should create tblx" -.amt 1.23`
	res, err := RunStmt(db, s)
	errFail(t, err)

	// Verify that node was added
	s = `find --silent -tbl tblx -select id,title,amt,body`
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Errorf("expected: 1 node, actual: %d node", len(res.Nodes))
	}

	// Edit node on nonexistent table should create the table
	s = `edit -tbl tbly -id 123 -.title "Should create tbly" -.amt 1.23`
	res, err = RunStmt(db, s)
	errFail(t, err)

	// Verify that table was created (find on table should not throw error)
	s = `find --silent -tbl tbly -select id,title,amt,body`
	_, err = RunStmt(db, s)
	errFail(t, err)

	// Find on nonexistent table should throw error
	s = `find --silent -tbl tblz -select id,title,amt,body`
	_, err = RunStmt(db, s)
	if err == nil {
		t.Errorf("find on nonexistent table should throw error")
	}
}

func TestFindSelect(t *testing.T) {
	fmt.Println("TestFindSelect...")

	s := `add -tbl t1 -.title "FindSelect test" -.amt 100.23 -.memo "select fields" "body text for find select"`
	_, err := RunStmt(db, s)

	// Find without specifying select (should return all cols)
	s = `find --silent -tbl t1 -w "title = 'FindSelect test' and amt = 100.23"`
	res, err := RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	n := res.Nodes[0]

	if _, ok := n["body"]; !ok {
		t.Errorf("expected: body field returned when no select specified")
	}
	if _, ok := n["memo"]; !ok {
		t.Errorf("expected: memo field returned when no select specified")
	}
	if n["body"] != "body text for find select\n" || n["memo"] != "select fields" {
		t.Errorf("incorrect body or memo fields returned. body: '%s', memo: '%s'", n["body"], n["memo"])
	}

	// Find specifying select fields (should return only those cols)
	s = `find --silent -tbl t1 -select amt -where "title = 'FindSelect test' and amt = 100.23"`
	res, err = RunStmt(db, s)
	errFail(t, err)
	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}
	if res.SumNode != nil {
		t.Fatalf("expected nil SumNode when -total not specified")
	}

	n = res.Nodes[0]

	if _, ok := n["id"]; ok {
		t.Errorf("expected: id field should not be returned since not specified in select")
	}
	if _, ok := n["body"]; ok {
		t.Errorf("expected: body field should not be returned since not specified in select")
	}
	if _, ok := n["memo"]; ok {
		t.Errorf("expected: memo field should not be returned since not specified in select")
	}
	if _, ok := n["amt"]; !ok {
		t.Errorf("expected: amt field should be returned since not specified in select")
	}
	if n["amt"] != 100.23 {
		t.Errorf("expected amt: %f, actual amt: %f", 100.23, n["amt"])
	}
}

func TestSpecialCaseParms(t *testing.T) {
	fmt.Println("TestSpecialCaseParms...")

	s := `add -tbl t1 -.title "special case 1" "body line 1" "body line 2" -- "-body line 3 starting with dash"`
	res, err := RunStmt(db, s)
	errFail(t, err)

	s = `find --silent -tbl t1 -w "title = 'special case 1'`
	res, err = RunStmt(db, s)
	errFail(t, err)

	if len(res.Nodes) != 1 {
		t.Fatalf("expected: 1 node, actual: %d node", len(res.Nodes))
	}

	n := res.Nodes[0]
	if n["body"] != "body line 1\nbody line 2\n-body line 3 starting with dash\n" {
		t.Errorf("expected body should be 3 lines, actual body:\n%s", n["body"])
	}
}

func TestInfoTables(t *testing.T) {
	fmt.Println("TestInfoTables...")

	PurgeTable(db, "t1")
	PurgeTable(db, "tblx")
	PurgeTable(db, "tbly")

  // Dummy command to make sure nbdata table is created
	s := `edit --silent -id nonexistingid`
	res, err := RunStmt(db, s)

	s = `info --silent tables`
	res, err = RunStmt(db, s)
	if len(res.Infos) != 1 {
		t.Fatalf("expected info tables returns 1 table (nbdata) only")
	}
	if res.Infos[0] != "nbdata" {
		t.Fatalf("expected info tables returns 1 table (nbdata) only")
	}

	// Add to new tables
	s = `add --silent -tbl t1 -.title "something" "body something"`
	res, err = RunStmt(db, s)
	errFail(t, err)
	s = `add --silent -tbl tblx -.title "something" "body something"`
	res, err = RunStmt(db, s)
	errFail(t, err)
	s = `add --silent -tbl tbly -.title "something" "body something"`
	res, err = RunStmt(db, s)
	errFail(t, err)

	s = `info --silent tables`
	res, err = RunStmt(db, s)
	if len(res.Infos) != 4 {
		t.Fatalf("expected info tables returns 4 tables")
	}
	if res.Infos[0] != "nbdata" {
		t.Fatalf("expected info tables returns nbdata in list")
	}
	if res.Infos[1] != "t1" {
		t.Fatalf("expected info tables returns t1 in list")
	}
	if res.Infos[2] != "tblx" {
		t.Fatalf("expected info tables returns tblx in list")
	}
	if res.Infos[3] != "tbly" {
		t.Fatalf("expected info tables returns tbly in list")
	}
}

func TestInfoFields(t *testing.T) {
	fmt.Println("TestInfoFields...")

	PurgeTable(db, "tblx")

	// Add new fields
	s := `add --silent -tbl tblx -.title "something" "body something" -.amt 123.45`
	res, err := RunStmt(db, s)
	errFail(t, err)
	s = `add --silent -tbl tblx -.title "something" "body something" -.cat cat1,cat2,cat3`
	res, err = RunStmt(db, s)
	errFail(t, err)

	s = `info --silent fields`
	res, err = RunStmt(db, s)
	errFail(t, err)

	s = `info --silent fields tblx`
	res, err = RunStmt(db, s)
	errFail(t, err)

	fields := res.Infos
	if len(fields) != 7 {
		t.Fatalf("info fields tblx must return 7 fields")
	}
	if fields[0] != "id" {
		t.Errorf("info fields tblx must return 'id' (in right order)")
	}
	if fields[1] != "pubdate" {
		t.Errorf("info fields tblx must return 'pubdate' (in right order)")
	}
	if fields[2] != "moddate" {
		t.Errorf("info fields tblx must return 'moddate' (in right order)")
	}
	if fields[3] != "title" {
		t.Errorf("info fields tblx must return 'title' (in right order)")
	}
	if fields[4] != "body" {
		t.Errorf("info fields tblx must return 'body' (in right order)")
	}
	if fields[5] != "amt" {
		t.Errorf("info fields tblx must return 'amt' (in right order)")
	}
	if fields[6] != "cat" {
		t.Errorf("info fields tblx must return 'cat' (in right order)")
	}

	s = `info --silent fields tblz`
	res, err = RunStmt(db, s)
	errFail(t, err)

	fields = res.Infos
	if len(fields) != 0 {
		t.Fatalf("info fields tblz (nonexisting table) must return 0 fields")
	}
}
