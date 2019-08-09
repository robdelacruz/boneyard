package core

import (
	"bytes"
	"e3/cmdutil"
	"e3/store"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

func RunPipelineStmts(scmd string, r io.Reader, w io.Writer, st *store.Store, opts, aliases map[string]string, logger *log.Logger) {
	e3c := E3C{st, opts, aliases, logger}
	jt := cmdutil.NewJumpTbl()
	jt.Handle("createdb", e3c.Createdb)
	jt.Handle("new", e3c.New)
	jt.Handle("load", e3c.Load)
	jt.Handle("update", e3c.Update)
	jt.Handle("search", e3c.Search)
	jt.Handle("find", e3c.Find)
	jt.Handle("map", e3c.Map)
	jt.Handle("edit", e3c.Edit)
	jt.Handle("echo", e3c.Echo)
	jt.Handle("bgindex", e3c.BgIndex)

	//	aliases := map[string]string{
	//		"assignto":  "map -assigned=$1, update",
	//		"showtable": "map -inputfmt=$1 -outputfmt=table",
	//		"convert":   "map -inputfmt=$1 -outputfmt=$2",
	//		"s":         "search",
	//		"listnodes": "load -outputfmt=table",
	//	}

	var out io.ReadWriter
	in := r
	if in == nil {
		in = os.Stdin
	}

	stmts := parsePipelineStmts(scmd)

	// Expand aliases until no more aliases to expand
	wasExpanded := true
	recurseLimit := 30
	for {
		stmts, wasExpanded = expandAliases(stmts, aliases)
		if !wasExpanded {
			break
		}
		recurseLimit--
		if recurseLimit <= 0 {
			break
		}
	}

	for _, stmt := range stmts {
		var b bytes.Buffer
		out = &b

		resp, err := execStmt(stmt, jt, in, out, opts)
		if err != nil {
			fmt.Println("error(s) occured, check logs.")
			s := fmt.Sprintf("error running statement '%s' (%s)", stmt, err)
			logger.Fatalf(s)
		}

		logger.Printf("$ %s\n", stmt)
		logger.Printf("> %s\n", resp)

		in = out
	}

	if w == nil {
		w = os.Stdout
	}
	io.Copy(w, out)
}

func execStmt(stmt string, jt cmdutil.JumpTbl, r io.Reader, w io.Writer, opts map[string]string) (*cmdutil.Resp, error) {
	verb, args, nargs := parseStmt(stmt)

	req := &cmdutil.Req{
		Opts:  opts,
		Args:  args,
		Nargs: nargs,
	}

	return jt.Exec(verb, req, r, w)
}

func parsePipelineStmts(scmd string) []string {
	var stmts []string
	var stmt string

	openQuote := ' '
	escMode := false

	for _, c := range scmd {
		if escMode {
			stmt += string(c)
			escMode = false
			continue
		}

		// In quote mode
		if openQuote == '"' || openQuote == '\'' {
			if c == '\\' {
				// Begin escape mode within quotes
				escMode = true
			} else if c == openQuote {
				// End quote mode
				stmt += string(c)
				openQuote = ' '
			} else {
				stmt += string(c)
			}
			continue
		}

		// Skip whitespace at start of statement
		if stmt == "" && unicode.IsSpace(c) {
			continue
		}

		// Begin escape mode
		if c == '\\' {
			escMode = true
			continue
		}

		// Begin quote mode
		if c == '"' || c == '\'' {
			stmt += string(c)
			openQuote = c
			continue
		}

		// End statement in pipeline
		if c == ',' {
			stmts = append(stmts, stmt)
			stmt = ""
			continue
		}

		stmt += string(c)
	}

	if stmt != "" {
		stmts = append(stmts, stmt)
		stmt = ""
	}

	return stmts
}

func parseStmt(stmt string) (string, []string, map[string]string) {
	var verb string
	var args []string
	nargs := map[string]string{}

	runes := []rune(stmt)
	i := 0

	// Get verb (first word)
	for ; i < len(runes); i++ {
		c := runes[i]

		// Skip over leading whitespace
		if verb == "" && unicode.IsSpace(c) {
			continue
		}

		// End word
		if unicode.IsSpace(c) {
			break
		}

		verb += string(c)
	}

	i++

	openQuote := ' '
	prefixChar := ' '
	arg := ""

	nargK := ""
	nargV := ""
	nargKSet := ""

	// Get args and named args
	// Given: verb arg1 arg2 -narg1=val1 -narg2=val2 notnarg=notval
	// args: arg1, arg2, notarg=notval
	// nargs: nargs["narg1"] = val1, nargs["narg2"] = val2
	for ; i < len(runes); i++ {
		c := runes[i]

		// Defining an narg key -nargkey=val
		if nargK != "" {
			if unicode.IsSpace(c) {
				// -nargkey by itself is same as -nargkey=""
				nargs[nargK] = ""
				nargK = ""
			} else if c == '=' {
				nargKSet = nargK
				nargK = ""
			} else {
				nargK += string(c)
			}
			continue
		}

		if nargKSet != "" {
			if openQuote != ' ' {
				// -nargkey="val"
				if c == openQuote {
					nargs[nargKSet] = nargV
					nargKSet = ""
					nargV = ""

					openQuote = ' '
				} else {
					nargV += string(c)
				}
			} else if c == '"' || c == '\'' {
				openQuote = c
			} else if unicode.IsSpace(c) {
				// -nargkey=val
				nargs[nargKSet] = nargV
				nargKSet = ""
				nargV = ""
			} else {
				nargV += string(c)
			}
			continue
		}

		// Add narg key chars (-nargk...)
		if prefixChar == '-' {
			if unicode.IsSpace(c) {
				args = append(args, string(prefixChar))
			} else {
				nargK += string(c)
			}
			prefixChar = ' '
			continue
		}

		// In arg quote mode,
		// args are in quotes "arg1" "arg2"
		if openQuote != ' ' {
			if c == openQuote {
				// End quote mode, add "arg1"
				args = append(args, arg)
				arg = ""
				openQuote = ' '
			} else {
				arg += string(c)
			}
			continue
		}

		// In arg mode, unquoted arg1, arg2
		if arg != "" {
			if unicode.IsSpace(c) {
				args = append(args, arg)
				arg = ""
			} else {
				arg += string(c)
			}
			continue
		}

		// Begin quote mode, "arg1" "arg2" in quotes
		if c == '"' || c == '\'' {
			openQuote = c
			continue
		}

		// Begin -nargskey=val mode
		if c == '-' {
			prefixChar = c
			continue
		}

		if unicode.IsSpace(c) {
			continue
		}

		// Begin arg mode
		arg = string(c)
	}

	// Add any rightmost arg or -nargkey=val
	if arg != "" {
		args = append(args, arg)
		arg = ""
	}
	if nargK != "" {
		nargKSet = nargK
	}
	if nargKSet != "" {
		nargs[nargKSet] = nargV
		nargKSet = ""
		nargV = ""
	}

	return verb, args, nargs
}

func quoteArgs(args []string) []string {
	for i, arg := range args {
		args[i] = fmt.Sprintf("\"%s\"", arg)
	}
	return args
}

func expandAliases(stmts []string, aliases map[string]string) ([]string, bool) {
	retStmts := []string{}
	wasExpanded := false

	for _, stmt := range stmts {
		verb, args, _ := parseStmt(stmt)

		aliasCmd := aliases[verb]
		if aliasCmd == "" {
			retStmts = append(retStmts, stmt)
			continue
		}

		wasExpanded = true
		aliasCmd, unrefArgs := expandAliasArgs(aliasCmd, args)
		aliasStmts := parsePipelineStmts(aliasCmd)

		// First alias stmt is passed any unused (unreplaced) args
		if len(aliasStmts) > 0 {
			aliasStmts[0] += " " + strings.Join(quoteArgs(unrefArgs), " ")
		}

		for _, aliasStmt := range aliasStmts {
			retStmts = append(retStmts, aliasStmt)
		}
	}

	return retStmts, wasExpanded
}

// Given:
// cmd = pipeline cmd
// args = args list
//
// Replace all $n in cmd with args[$n] where n is an int from 1 to 9.
// Return transformed cmd and remaining unreferenced args.
// Ex.
// expandAliasArgs("load $1 $2 $3 -outputfmt=$4", ["a", "b", "c", "d", "e"]
// Returns:
//   "load a b c -outputfmt=d", ["e"]
func expandAliasArgs(scmd string, args []string) (string, []string) {
	expandedCmd := ""
	paramMode := false
	usedArgs := map[int]bool{}

	for _, c := range scmd {
		// In $n param mode
		if paramMode {
			// Replace $n with args[n-1]
			if c >= '1' && c <= '9' {
				i := int(c - '1')
				if len(args) > i {
					expandedCmd += args[i]
					usedArgs[i] = true
				}
			}
			paramMode = false
			continue
		}

		// Begin $n param mode
		if c == '$' {
			paramMode = true
			continue
		}

		expandedCmd += string(c)
	}

	unrefArgs := []string{}
	for i, _ := range args {
		if !usedArgs[i] {
			unrefArgs = append(unrefArgs, args[i])
		}
	}

	return expandedCmd, unrefArgs
}
