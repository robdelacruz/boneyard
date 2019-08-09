package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Node map[string]string

var _rxpID, _rxpField, _rxpEmptyField *regexp.Regexp
var _fieldTypes map[string]string
var _fieldWidths map[string]string

func init() {
	// ID line regex
	// %% id123
	_rxpID = regexp.MustCompile(`^%%\s+(\S+)\s*$`)

	// field line regex
	// .field val
	_rxpField = regexp.MustCompile(`^\.(\S+)\s+(.*)\s*$`)

	// empty field regex
	// .field
	_rxpEmptyField = regexp.MustCompile(`^\.(\S+)$`)

	initFieldSettings()

}

func initFieldSettings() {
	// _fieldTypes defines the type of the field:
	// f = float
	// d = int
	// s = string (default if no type specified)

	// _fieldWidths defines the print width format of the field:
	// This is used in --table output when writing fields in columns.
	// <n> = reserve at least n digits for 'd' (int) type
	// <w> = reserve at least w chars for 's' (string) type
	// <nnn.dd> = reserve at least nnn.dd number and decimal places
	// -<width> = for strings: right-justify
	//            for numbers: left-justify
	//            Ex. -3, -8.2

	// Set up defaults for standard field names
	_fieldTypes = map[string]string{
		"debit":  "f",
		"credit": "f",
		"price":  "f",
		"weight": "f",
		"age":    "d",
	}

	_fieldWidths = map[string]string{
		"id":     "20",
		"title":  "20",
		"body":   "15",
		"tags":   "10",
		"debit":  "8.2",
		"credit": "8.2",
		"price":  "8.2",
		"weight": "5.2",
		"age":    "-3",
	}

	// Load additional field settings from conf file
	nbconf := os.Getenv("HOME") + "/.nbconf"
	if _, err := os.Stat(nbconf); os.IsNotExist(err) {
		return // no .nbconf settings
	}
	f, err := os.Open(nbconf)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	parseNBData(f, func(node Node) {
		if node["id"] == "fieldtypes" {
			for k, v := range node {
				if k == "id" || k == "body" {
					continue
				}
				if k == "-body" {
					k = "body"
				}
				_fieldTypes[k] = strings.TrimSpace(v)
			}
			return
		}
		if node["id"] == "fieldwidths" {
			for k, v := range node {
				if k == "id" || k == "body" {
					continue
				}
				if k == "-body" {
					k = "body"
				}
				_fieldWidths[k] = strings.TrimSpace(v)
			}
			return
		}
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, `Notes Buddy is a tool for managing notes.
Usage:
  nb command [arguments]

Commands:
  help   Show this help screen
  add    Add note
	patch  Patch note
	clear  Clear notes
  del    Delete notes
  find   Find notes
`)
		os.Exit(0)
	}

	cmd := os.Args[1]

	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	switches, parms := parseArgs(args)

	if cmd == "find" {
		find(switches, parms)
	} else if cmd == "add" {
		add(switches, parms)
	} else if cmd == "patch" {
		patch(switches, parms)
	} else if cmd == "clear" {
		clear(switches, parms)
	} else if cmd == "del" {
		del(switches, parms)
	}

}

// Read switches and standalone parameters from an arg list.
// Ex. -author robtwister -title "Note Title" -flag1 -ytd "Body text here"
// returns -
// switches: [{author=robtwister}, {title="Note Title"}, {flag1=""}, {ytd=""}]
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
		} else if strings.HasPrefix(arg, "-") {
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

// Read data file and apply filter to nodes
func parseNBData(f io.Reader, filterFunc func(node Node)) {
	nodeTextLines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := _rxpID.FindStringSubmatch(line); matches != nil {
			// Parse node lines and add new node to list
			if len(nodeTextLines) > 0 {
				f := bytes.NewBufferString(strings.Join(nodeTextLines, "\n"))
				node := parseNodeText(f)
				if filterFunc != nil {
					filterFunc(node)
				}
			}

			// Reset node lines
			nodeTextLines = []string{}
		}

		nodeTextLines = append(nodeTextLines, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error(%s)\n", err)
	}

	// Parse remaining node lines and add to nodes list
	if len(nodeTextLines) > 0 {
		f := strings.NewReader(strings.Join(nodeTextLines, "\n"))
		node := parseNodeText(f)
		if filterFunc != nil {
			filterFunc(node)
		}
	}
}

func parseNodeText(f io.Reader) Node {
	// Parse stdin lines into node struct
	node := Node{}
	bodyLines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := _rxpID.FindStringSubmatch(line); matches != nil {
			// ID line
			id := matches[1]
			node["id"] = id
		} else if matches = _rxpField.FindStringSubmatch(line); matches != nil {
			// Field line - set field value
			field := matches[1]
			val := matches[2]
			node[field] = val
		} else if matches = _rxpEmptyField.FindStringSubmatch(line); matches != nil {
			// Empty field line - delete previous field definition
			field := matches[1]
			delete(node, field)
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

func writeNodeCardField(f io.Writer, fieldName, fieldVal string, displayFields []string) {
	if displayFields == nil || inSlc(displayFields, fieldName) {
		fmt.Fprintf(f, ".%s %s\n", fieldName, fieldVal)
	}
}

// Write node in card format.
// %% <id>
// .field1 <val>
// .field2  <val>
// <body text>
func writeNodeCard(f io.Writer, node Node, fields []string) {
	if fields == nil || inSlc(fields, "id") {
		fmt.Fprintf(f, "%%%% %s\n", node["id"])
	}
	if title, ok := node["title"]; ok {
		writeNodeCardField(f, "title", title, fields)
	}

	for field, val := range node {
		if field == "id" || field == "body" || field == "title" {
			continue
		}
		writeNodeCardField(f, field, val, fields)
	}

	if body, ok := node["body"]; ok {
		if fields == nil || inSlc(fields, "body") {
			fmt.Fprintf(f, "%s\n", body)
		}
	}
}

func writeNodeRow(f io.Writer, node Node, fields []string) {
	for _, field := range fields {
		fieldType := _fieldTypes[field]
		if fieldType == "" {
			fieldType = "s"
		}

		var fieldFmt string
		fieldWidth := _fieldWidths[field]
		if fieldWidth != "" {
			if fieldType == "s" {
				// "%-20.20s" where width=20
				fieldFmt = fmt.Sprintf("%%-%s.%ss", fieldWidth, fieldWidth)
			} else {
				// "%8.2f" where width=8.2 and fieldType=f
				fieldFmt = fmt.Sprintf("%%%s%s", fieldWidth, fieldType)
			}
		} else {
			// Default format specs
			if fieldType == "f" {
				fieldFmt = "%5.2f"
			} else if fieldType == "d" {
				fieldFmt = "%5d"
			} else {
				fieldFmt = "%-10.10s"
			}
		}

		v := node[field]

		if fieldType == "f" {
			n, _ := strconv.ParseFloat(v, 64)
			fmt.Fprintf(f, fieldFmt, n)
		} else if fieldType == "d" {
			n, _ := strconv.ParseInt(v, 10, 32)
			fmt.Fprintf(f, fieldFmt, n)
		} else {
			fmt.Fprintf(f, fieldFmt, v)
		}
		fmt.Fprintf(f, "  ")
	}
	fmt.Fprintf(f, "\n")
}

func updateTotals(node Node, totals map[string]float64) {
	for k, v := range node {
		if k == "id" || k == "body" {
			continue
		}
		fieldType := _fieldTypes[k]
		if fieldType == "" || fieldType == "s" {
			continue
		}
		if fieldType == "f" {
			n, _ := strconv.ParseFloat(v, 64)
			totals[k] += n
			continue
		}
		if fieldType == "d" {
			n, _ := strconv.ParseInt(v, 10, 32)
			totals[k] += float64(n)
		}
	}
}

func find(switches map[string]string, parms []string) {
	var fields []string
	if switches["fields"] != "" {
		fields = strings.Split(switches["fields"], ",")
	}

	totals := map[string]float64{}

	nodes := findNodes(switches, parms)
	for _, node := range nodes {
		if switches["table"] != "" {
			writeNodeRow(os.Stdout, node, fields)
		} else {
			writeNodeCard(os.Stdout, node, fields)
		}

		if switches["totals"] != "" {
			updateTotals(node, totals)
		}
	}

	if switches["totals"] != "" {
		for k, v := range totals {
			fmt.Printf("totals[%s] = %f\n", k, v)
		}
	}
}

func findNodes(switches map[string]string, parms []string) []Node {
	// Open data file
	nbfile := os.Getenv("HOME") + "/.nbdata"
	f, err := os.Open(nbfile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var ep *ExprParser
	if switches["q"] != "" {
		buf := bytes.NewBufferString(switches["q"])
		ep = NewExprParser(buf)
	}

	nodes := []Node{}
	parseNBData(f, func(node Node) {
		if matchNode(switches, parms, ep, node) {
			nodes = append(nodes, node)
		}
	})
	return nodes
}

// Switches:
// --all         match all nodes
// -id <id>      exact match on id
// -q <where>    where condition
//
// <where> is composed of multiple condition clauses that can be grouped
// together in parentheses and combined with 'and', 'or'.
//
// <where> takes the following form:
// <field> <op> <val>
//
// where <op> can be either of the following:
// =              (exact match)
// <>             (not exact match)
// =~             (regex match)
// >, >=, <, <=   (comparison operators)
// =~             (regex match)
//
// Ex.
// -q "(cat='commute' or cat='dineout') and date>='2018-08-01'"
//
// Parms:
// <parm1> <parm2>  partial match any field containing <parm1> or <parm2>
//                  excluding fields 'id', 'date'
//
func matchNode(switches map[string]string, parms []string, ep *ExprParser, node Node) bool {
	// -id <id>  match specific id
	if switches["id"] != "" {
		if node["id"] == switches["id"] {
			return true
		}
		return false
	}
	// --all matches all nodes
	if switches["all"] != "" {
		// Show all nodes
		return true
	}

	// -q <condition expression>
	if ep != nil {
		env := Env{
			Vars:     node,
			VarTypes: _fieldTypes,
		}
		ep.Reset()
		opr, err := ep.Run(&env)
		if err != nil {
			return false
		}
		if opr.Val != "0" {
			return true
		}
		return false
	}

	// regex search for each parm on each field
	if len(parms) > 0 {
		for _, q := range parms {
			rxpMatch := regexp.MustCompile("(?i)" + q)
			for field, val := range node {
				// Don't match on id or date field
				if field == "id" || field == "date" {
					continue
				}
				if rxpMatch.MatchString(val) {
					return true
				}
			}
		}
		// no match
		return false
	}

	// where condition
	if switches["q"] != "" {
	}

	// no match
	return false
}

func add(switches map[string]string, parms []string) {
	node := Node{}

	if len(parms) > 0 {
		// Body text specified, use this in node
		body := strings.Join(parms, "\n")
		node["body"] = body
	} else {
		// Print instructions
		fmt.Println(`Sample:
-----------------------------------
.title New Note
.author rob
.tags tag1, multi-word tag, tag2
.custom add as many custom fields
This is the first paragraph of the
note.

Add as many lines and paragraphs as
needed in the note.
-----------------------------------

Enter new note below (CTRL-D to end):
`)

		// No body text specified, read body and fields from stdin
		node = parseNodeText(os.Stdin)
	}

	// Add any specified field definition.
	// Field definitions are switches starting with a dot (.)
	// Ex. -.title "Note Title" -.author "rob"
	// Fields defined from stdin takes precedence over cli fields.
	for k, v := range switches {
		if strings.HasPrefix(k, ".") {
			field := k[1:]
			if node[field] == "" {
				node[field] = v
			}
		}
	}

	// Auto-generate ID
	node["id"] = GenID()

	// Open data file
	nbfile := os.Getenv("HOME") + "/.nbdata"
	f, err := os.OpenFile(nbfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writeNodeCard(f, node, nil)
}

func patch(switches map[string]string, parms []string) {
	sIDs := switches["id"]
	if sIDs == "" {
		fmt.Fprintf(os.Stderr, `Please specify a node id to patch.

Example:
$ nb patch -id -LKHBOXNwtl1rdx-PmF- "Line to add"

`)
		return
	}

	argIDs := strings.Split(sIDs, " ")

	addLine := strings.Join(parms, "\n")
	if len(addLine) == 0 {
		fmt.Fprintf(os.Stderr, `Please specify what to add to the node.

Example:
$ nb patch -id -LKHBOXNwtl1rdx-PmF- "Line to add"

`)
		return
	}

	// Open data file
	nbfile := os.Getenv("HOME") + "/.nbdata"
	f, err := os.Open(nbfile)
	if err != nil {
		panic(err)
	}

	// Open file to write to
	nbfileOut := os.Getenv("HOME") + "/.nbdata.out"
	fOut, err := os.OpenFile(nbfileOut, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	fInNode := false
	nodeLines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Start of new node
		if matches := _rxpID.FindStringSubmatch(line); matches != nil {
			// If previous node is target of patch,
			// Add patch line and write previous node.
			if fInNode {
				nodeLines = append(nodeLines, addLine)
				fNode := bytes.NewBufferString(strings.Join(nodeLines, "\n"))
				writeNodeCard(fOut, parseNodeText(fNode), nil)

				nodeLines = []string{} // reset node lines
			}

			id := matches[1]
			// Mark whether this is target node to patch
			if inSlc(argIDs, id) {
				fInNode = true
			} else {
				fInNode = false
			}
		}

		if fInNode {
			nodeLines = append(nodeLines, line)
		} else {
			fmt.Fprintf(fOut, "%s\n", line)
		}
	}

	// Add patch line and write previous target node if any.
	if fInNode {
		nodeLines = append(nodeLines, addLine)
		fNode := bytes.NewBufferString(strings.Join(nodeLines, "\n"))
		writeNodeCard(fOut, parseNodeText(fNode), nil)

		nodeLines = []string{} // reset node lines
	}

	f.Close()
	fOut.Close()

	// nbdata.out -> nbdata
	os.Rename(nbfile, nbfile+".bak")
	os.Rename(nbfileOut, nbfile)
}

func clear(switches map[string]string, parms []string) {
	sIDs := switches["id"]
	if sIDs == "" {
		fmt.Fprintf(os.Stderr, `Please specify a node id to clear.

Example:
$ nb clear -id -LKHBOXNwtl1rdx-PmF-

`)
		return
	}

	argIDs := strings.Split(sIDs, " ")

	// Open data file
	nbfile := os.Getenv("HOME") + "/.nbdata"
	f, err := os.Open(nbfile)
	if err != nil {
		panic(err)
	}

	// Open file to write to
	nbfileOut := os.Getenv("HOME") + "/.nbdata.out"
	fOut, err := os.OpenFile(nbfileOut, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	fInNode := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := _rxpID.FindStringSubmatch(line); matches != nil {
			id := matches[1]
			// Mark whether line is currently part of node to clear
			if inSlc(argIDs, id) {
				fInNode = true
			} else {
				fInNode = false
			}

			fmt.Fprintf(fOut, "%s\n", line)
			continue
		}

		// Don't write inner lines of node to clear
		if !fInNode {
			fmt.Fprintf(fOut, "%s\n", line)
		}
	}

	f.Close()
	fOut.Close()

	// nbdata.out -> nbdata
	os.Rename(nbfile, nbfile+".bak")
	os.Rename(nbfileOut, nbfile)
}

func del(switches map[string]string, parms []string) {
	sIDs := switches["id"]
	if sIDs == "" {
		fmt.Fprintf(os.Stderr, `Please specify a node id to delete.

Example:
$ nb del -id -LKHBOXNwtl1rdx-PmF-

`)
		return
	}

	argIDs := strings.Split(sIDs, " ")

	// Open data file
	nbfile := os.Getenv("HOME") + "/.nbdata"
	f, err := os.Open(nbfile)
	if err != nil {
		panic(err)
	}

	// Open file to write to
	nbfileOut := os.Getenv("HOME") + "/.nbdata.out"
	fOut, err := os.OpenFile(nbfileOut, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	fInNode := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := _rxpID.FindStringSubmatch(line); matches != nil {
			id := matches[1]
			// Mark whether line is currently part of node to delete
			if inSlc(argIDs, id) {
				fInNode = true
			} else {
				fInNode = false
			}
		}

		// Don't write lines part of node to delete
		if !fInNode {
			fmt.Fprintf(fOut, "%s\n", line)
		}
	}

	f.Close()
	fOut.Close()

	// nbdata.out -> nbdata
	os.Rename(nbfile, nbfile+".bak")
	os.Rename(nbfileOut, nbfile)
}
