package main

import (
	"e3/cmdutil"
	"e3/core"
	"e3/store"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage:\n")
		os.Exit(0)
	}

	opts, aliases, args, err := readOptions()
	if err != nil {
		fmt.Printf("Unable to load initial settings (%s)\n", err)
		os.Exit(1)
	}

	dbdriver := opts["dbdriver"]
	if dbdriver == "" {
		fmt.Printf("Please define a database driver (dbdriver= in conf file), Ex. dbdriver=sqlite3\n")
		os.Exit(1)
	}

	dsname := opts["dsname"]
	if dsname == "" {
		fmt.Printf("Please define a database source (dsname= in conf file), Ex. dsname=/var/local/nodes.db\n")
		os.Exit(1)
	}

	indexDir := opts["indexdir"]
	if indexDir == "" {
		fmt.Printf("Please define an index dir (\"indexdir=\" in conf file)\n")
		os.Exit(1)
	}

	flog := os.Stdout
	logfile := opts["logfile"]
	if logfile != "" {
		flog, err = os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Can't open logfile %s (%s)", logfile, err)
			os.Exit(1)
		}
		defer flog.Close()
	}

	logger := log.New(flog, "", 0)
	st := store.NewStore(dbdriver, dsname, indexDir, logger)

	if cmdutil.FlagOn(opts, "http") {
		serveHttp(st, opts, aliases, logger)
		os.Exit(0)
	}

	// Run pipeline command by passing it as first arg or in -e switch:
	//   ./e -e "(pipeline commands)"
	//   ./e "(pipeline commands)"
	scmd := opts["eval"]
	if scmd == "" && len(args) > 0 {
		scmd = args[0]
	}
	if scmd == "" {
		os.Exit(1)
	}

	core.RunPipelineStmts(scmd, nil, nil, st, opts, aliases, logger)
}

func serveHttp(st *store.Store, opts, aliases map[string]string, logger *log.Logger) {
	e3c := core.NewE3C(st, opts, aliases, logger)

	http.HandleFunc("/", e3c.HttpRoot)
	http.HandleFunc("/cmd", e3c.HttpCmd)

	fmt.Println("Serving http at localhost:8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal(err)
	}
}
