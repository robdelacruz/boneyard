package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

const (
	MaxRowFieldChars = 35
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, `Notes Buddy is a tool for managing notes.
Usage:
  nb command [arguments]

Commands:
  help   Show this help screen
  add    Add node
  edit   Edit node
  del    Delete nodes
  find   Find nodes
  copy   Copy node
  info   Request info
  run    Run command
`)
		os.Exit(0)
	}

	db := InitDB()
	cmd := os.Args[1]

	var res CmdResult
	var err error

	// $ nb run "<script line>"
	if cmd == "run" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Nothing to run. (Ex. $ nb run \"<stmt>\")\n")
			os.Exit(0)
		}

		stmt := os.Args[2]
		res, err = RunStmt(db, stmt)
	} else {
		// $nb <cmd> <args>...
		var switches map[string]string
		var parms []string

		// $nb <cmd>
		if len(os.Args) > 2 {
			switches, parms = parseArgs(os.Args[2:])
		}
		res, err = RunCmd(db, cmd, switches, parms)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if SlcContains(os.Args, "--debug") {
		fmt.Println("*** Result ***")
		fmt.Print(&res)
		fmt.Println("*** End ***")
	}
	os.Exit(0)
}

type CmdResult struct {
	NodeID  string
	Node    Node
	Nodes   []Node
	SumNode Node
	Infos   []string
}

func (res *CmdResult) String() string {
	var b bytes.Buffer
	if res.NodeID != "" {
		b.WriteString(fmt.Sprintf("NodeID: %s\n", res.NodeID))
	}
	if res.Node != nil {
		b.WriteString("Node\n")
		b.WriteString("----\n")
		b.WriteString(fmt.Sprintf("%s\n", res.Node))
	}
	if res.Nodes != nil {
		b.WriteString("Nodes\n")
		b.WriteString("-----\n")
		for _, n := range res.Nodes {
			b.WriteString(fmt.Sprintf("%s", n))
			b.WriteString("-----\n")
		}
	}
	if res.SumNode != nil {
		b.WriteString("SumNode\n")
		b.WriteString("-------\n")
		b.WriteString(fmt.Sprintf("%s\n", res.SumNode))
	}
	return b.String()
}

func nodeHeaders(cols []string) map[string]string {
	hh := map[string]string{}
	for _, c := range cols {
		hh[c] = c
	}
	return hh
}

func writeHeaderRow(w io.Writer, hh map[string]string, cols []string, fldLens map[string]int) {
	for _, col := range cols {
		lenCol := fldLens[col]
		if (col == "body" || col == "title") && (lenCol > MaxRowFieldChars) {
			lenCol = MaxRowFieldChars
		}
		fmtspec := fmt.Sprintf("%%-%ds  ", lenCol)
		fmt.Fprintf(w, fmtspec, hh[col])
	}
	fmt.Fprintf(w, "\n")
}

func writeSepRow(w io.Writer, cols []string, fldLens map[string]int) {
	for _, col := range cols {
		lenCol := fldLens[col]
		if (col == "body" || col == "title") && (lenCol > MaxRowFieldChars) {
			lenCol = MaxRowFieldChars
		}
		dashes := strings.Repeat("-", lenCol)
		fmtspec := fmt.Sprintf("%%-%ds  ", lenCol)
		fmt.Fprintf(w, fmtspec, dashes)
	}
	fmt.Fprintf(w, "\n")
}

func writeNodeRow(w io.Writer, n Node, cols []string, fldLens map[string]int) {
	for _, col := range cols {
		rval := []rune(StrV(n[col]))

		lenCol := fldLens[col]
		// For body field, don't display newlines and clip length.
		if (col == "body" || col == "title") && (lenCol > MaxRowFieldChars) {
			lenCol = MaxRowFieldChars
		}

		// Show only first lenCol chars in body
		// Show "..." at the end if longer than lenCol chars.
		if len(rval) > lenCol {
			rval = rval[:lenCol-3]
			rval = append(rval, '.', '.', '.')
		}

		// Don't display newlines or tabs.
		sval := strings.Replace(string(rval), "\n", "", -1)
		sval = strings.Replace(sval, "\t", "", -1)

		fmtspec := fmt.Sprintf("%%-%d.%[1]ds  ", lenCol)
		fmt.Fprintf(w, fmtspec, sval)
	}
	fmt.Fprintf(w, "\n")
}

func writeNodeRecord(w io.Writer, n Node, cols []string) {
	for _, col := range cols {
		if col == "body" {
			continue
		}
		fmt.Fprintf(w, ".%s %s\n", col, StrV(n[col]))
	}

	if SlcContains(cols, "body") {
		fmt.Fprintf(w, "%s\n", StrV(n["body"]))
	}
}

// Read switches and standalone parameters from an arg list.
// Ex. -author robtwister -title "Note Title" -flag1 --ytd "Body text here"
// returns -
// switches: [{author=robtwister}, {title="Note Title"}, {flag1=""}, {ytd="y"}]
// parms: ["Body text here"]
//
// To determine whether a switch is standalone (Ex. -ytd), it is cross-checked
// with a list of reserved switch words.
func parseArgs(args []string) (switches map[string]string, parms []string) {
	switches = map[string]string{}
	parms = []string{}

	fNoMoreSwitches := false
	curKey := ""

	for _, arg := range args {
		if fNoMoreSwitches {
			// any arg after "--" is a standalone parameter
			parms = append(parms, arg)
		} else if arg == "--" {
			// "--" means no more switches to come
			fNoMoreSwitches = true
		} else if strings.HasPrefix(arg, "--") {
			switches[arg[2:]] = "y"
			curKey = ""
		} else if curKey == "-id" {
			// special case handling for "-id <id>"
			// (because <id> could start with "-" and be interpreted as switch)
			switches["id"] = arg
			curKey = ""
		} else if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "-=") {
			// -switchA
			curKey = arg
		} else if curKey != "" {
			switches[curKey[1:]] = arg // strip out "-"
			curKey = ""
		} else {
			// standalone parameter
			parms = append(parms, arg)
		}
	}

	return switches, parms
}

// Extract args from string cmd.
// Handles double/single quotes, escaped chars, like bash.
func parseCmd(cmd string) []string {
	var args []string
	var arg string
	escapeMode := false
	openQuoteChar := ' '

	for _, c := range cmd {
		if escapeMode {
			arg += string(c)
			escapeMode = false
			continue
		}
		// If within double or single quotes
		if openQuoteChar == '"' || openQuoteChar == '\'' {
			if c == '\\' {
				escapeMode = true
			} else if c == openQuoteChar {
				openQuoteChar = ' ' // ending quote
			} else {
				arg += string(c)
			}
			continue
		}
		if unicode.IsSpace(c) {
			if len(arg) > 0 {
				args = append(args, arg)
				arg = ""
			}
			continue
		}
		if c == '\\' {
			escapeMode = true
			continue
		}
		if c == '"' || c == '\'' {
			openQuoteChar = c
			continue
		}
		arg += string(c)
	}

	if len(arg) > 0 {
		args = append(args, arg)
	}
	return args
}

func RunStmt(db *sql.DB, stmt string) (CmdResult, error) {
	runargs := parseCmd(stmt)
	if len(runargs) == 0 {
		return CmdResult{}, nil
	}

	var switches map[string]string
	var parms []string

	cmd := runargs[0]
	if len(runargs) > 1 {
		switches, parms = parseArgs(runargs[1:])
	}

	return RunCmd(db, cmd, switches, parms)
}

func RunCmd(db *sql.DB, cmd string, switches map[string]string, parms []string) (CmdResult, error) {
	id := switches["id"]
	tbl := switches["tbl"]
	if tbl == "" {
		tbl = "nbdata"
	}

	var out io.Writer = os.Stdout
	if switches["silent"] != "" {
		var b bytes.Buffer
		out = &b
	}

	if cmd == "add" {
		InitTable(db, tbl)
		node := ReadNodeCLI(switches, parms)
		newID, err := AddNode(db, tbl, node)
		return CmdResult{NodeID: newID}, err

	} else if cmd == "edit" {
		if id == "" {
			fmt.Fprintf(out, `Usage:
nb edit -id <node id>
`)
			return CmdResult{}, nil
		}
		InitTable(db, tbl)
		node := ReadNodeCLI(switches, parms)
		err := UpdateNode(db, tbl, node, switches)
		return CmdResult{Node: node}, err

	} else if cmd == "del" {
		if id == "" {
			fmt.Fprintf(out, `Usage:
nb del -id <node id>
`)
			return CmdResult{}, nil
		}
		InitTable(db, tbl)
		err := DelNode(db, tbl, id)
		return CmdResult{}, err

	} else if cmd == "find" {
		// -select "col1,col2" -total "col2" -order "col1,col2" -where "col1 > 0.0"

		var sel, total, order, where string
		if switches["select"] != "" {
			sel = switches["select"]
		} else if switches["sel"] != "" {
			sel = switches["sel"]
		} else if switches["s"] != "" {
			sel = switches["s"]
		}
		if switches["total"] != "" {
			total = switches["total"]
		} else if switches["tot"] != "" {
			total = switches["tot"]
		} else if switches["t"] != "" {
			total = switches["t"]
		}
		if switches["order"] != "" {
			order = switches["order"]
		} else if switches["ord"] != "" {
			order = switches["ord"]
		} else if switches["o"] != "" {
			order = switches["o"]
		}
		if switches["where"] != "" {
			where = switches["where"]
		} else if switches["w"] != "" {
			where = switches["w"]
		}
		// -id <node id> to find single node with that unique id
		// It will override the where clause if one is specified.
		if id != "" {
			where = fmt.Sprintf("id = '%s'", id)
		}

		var selectCols, totalCols, orderCols []string
		r := regexp.MustCompile(`,\s*`)
		if sel != "" {
			selectCols = r.Split(sel, -1)
		}
		if total != "" {
			totalCols = r.Split(total, -1)
		}
		if order != "" {
			orderCols = r.Split(order, -1)
		}

		q := strings.Join(parms, " ")

		nn, selectCols, err := QueryNodes(db, tbl, selectCols, orderCols, where, q)
		if err != nil {
			return CmdResult{}, err
		}

		// Only include total cols that have corresponding select col.
		existingCols := []string{}
		for _, totalCol := range totalCols {
			if SlcContains(selectCols, totalCol) {
				existingCols = append(existingCols, totalCol)
			}
		}
		totalCols = existingCols

		// Query for Totals
		var sumNode Node
		if totalCols != nil && len(totalCols) > 0 {
			cc := []string{}
			for _, totalCol := range totalCols {
				cc = append(cc, fmt.Sprintf("SUM(%s) as %s", totalCol, totalCol))
			}
			totalNodes, _, err := QueryNodes(db, tbl, cc, nil, where, q)
			if err != nil {
				return CmdResult{Nodes: nn}, err
			}
			sumNode = totalNodes[0]
		}

		printNodes(out, nn, sumNode, selectCols, totalCols, switches)

		return CmdResult{Nodes: nn, SumNode: sumNode}, nil
	} else if cmd == "info" {
		if len(parms) == 0 {
			fmt.Fprintf(out, `Usage:
nb info <request>

Requests:
  tables           Show list of node tables
  fields <table>   Show fields of <table>
`)
			return CmdResult{}, nil
		}

		if parms[0] == "tables" {
			// ./nb info tables
			tt := NodeTables(db)
			for _, t := range tt {
				fmt.Fprintf(out, "%s\n", t)
			}
			return CmdResult{Infos: tt}, nil
		} else if parms[0] == "fields" {
			// ./nb info fields <tbl name>
			if len(parms) < 2 {
				fmt.Fprintf(out, `Usage:
nb info fields <table>
`)
				return CmdResult{}, nil
			}

			tbl := parms[1]
			cc := NodeTableCols(db, tbl)
			for _, c := range cc {
				fmt.Fprintf(out, "%s\n", c)
			}
			return CmdResult{Infos: cc}, nil
		}
	} else if cmd == "copy" {
		// -id <id>,...  to each node of <id>...
		// -w <where>    copy all nodes matching where clause
		//
		// nb copy -id 123,456 -tblsrc tblx -tbldest tbly
		// Copies tblx nodes '123', '456' to tbly.
		//
		// nb copy -w "cat like '%cat1,%'" -tblsrc tblx -tbldest tbly
		// Copies tblx where "cat like '%cat1,%'" to tbly.
		//
		// nb copy -id 123 -tbldest tbly
		// Copies nbdata (default table) '123' to tbly.
		//
		// nb copy -id 123 -tblsrc tblx
		// Copies tblx '123' to nbdata (default table).

		// INSERT INTO tbldest (id, title, body) FROM SELECT id, title, body FROM tblsrc
	}

	err := fmt.Errorf("Unknown command (%s)", cmd)
	return CmdResult{}, err
}

func maxFieldLengths(nn []Node) map[string]int {
	lens := map[string]int{}
	for _, n := range nn {
		// Keep track of longest field name or width length
		for k, v := range n {
			kLen := len([]rune(k))
			if kLen > lens[k] {
				lens[k] = kLen
			}

			vLen := len([]rune(StrV(v)))
			if vLen > lens[k] {
				lens[k] = vLen
			}
		}
	}
	return lens
}

func printNodes(w io.Writer, nn []Node, sumNode Node, selectCols []string, totalCols []string, switches map[string]string) {
	if switches["mode"] == "column" {
		// Print nodes in table format:
		// one row for each node, with one column for each field.

		// Iterate through nodes and get the longest char length of each field.
		fldLens := maxFieldLengths(nn)

		// Header and separator
		writeHeaderRow(w, nodeHeaders(selectCols), selectCols, fldLens)
		writeSepRow(w, selectCols, fldLens)

		// Each node is printed in a new row
		for _, n := range nn {
			writeNodeRow(w, n, selectCols, fldLens)
		}

		// Totals separator and row
		if sumNode != nil {
			writeSepRow(w, selectCols, fldLens)
			writeNodeRow(w, sumNode, selectCols, fldLens)
		}
	} else {
		// Print nodes in list format:
		// one row for each field and value in this format: field1: field val
		// except for the body field which will appear the next line after all the
		// other field/value rows. Body field is printed as is in multiple lines.
		// A '%%' line is printed in between node records.

		for _, n := range nn {
			writeNodeRecord(w, n, selectCols)
			fmt.Fprintln(w, "%%")
		}

		// Totals node
		if sumNode != nil {
			fmt.Fprintln(w, "*** Totals ***")
			writeNodeRecord(w, sumNode, totalCols)
		}
	}
}
