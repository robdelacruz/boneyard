package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	os.Args = os.Args[1:]
	switches, parms := parseArgs(os.Args)

	db, err := openDB(switches)
	if err != nil {
		log.Fatal(err)
	}

	cmd := "serve"
	if len(parms) > 0 {
		if parms[0] == "serve" || parms[0] == "info" || parms[0] == "help" {
			cmd = parms[0]
			parms = parms[1:]
		}
	}

	switch cmd {
	case "serve":
		port := "8000"
		if len(parms) > 0 {
			port = parms[0]
		}

		http.HandleFunc("/", rootHandler(db))
		http.HandleFunc("/task/", taskHandler(db))

		fmt.Printf("Listening on %s...\n", port)
		err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
		log.Fatal(err)
	}
}

func parseArgs(args []string) (map[string]string, []string) {
	switches := map[string]string{}
	parms := []string{}

	standaloneSwitches := []string{}
	definitionSwitches := []string{"F"}
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
		} else if strings.HasPrefix(arg, "-") {
			if listContains(definitionSwitches, arg[1:]) {
				// -a "val"
				curKey = arg[1:]
				continue
			}
			for _, ch := range arg[1:] {
				// -a, -b, -ab
				sch := string(ch)
				if listContains(standaloneSwitches, sch) {
					switches[sch] = "y"
				}
			}
		} else if curKey != "" {
			switches[curKey] = arg
			curKey = ""
		} else {
			// standalone parameter
			parms = append(parms, arg)
		}
	}

	return switches, parms
}

func listContains(ss []string, v string) bool {
	for _, s := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func rootHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from all sites.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.WriteHeader(http.StatusBadRequest)
	}
}

func taskHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var t Task
		switch r.Method {
		case "POST":
			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				log.Printf("taskHandler() json decoding error:\n%s\n", err)
				http.Error(w, "Invalid task.", http.StatusBadRequest)
				return
			}

			if t.Status < 0 || t.Status > 2 {
				t.Status = 0
			}
			if strings.TrimSpace(t.Createdt) == "" {
				t.Createdt = time.Now().Format(time.RFC3339)
			}

			err = insertTask(db, &t)
			if err != nil {
				log.Printf("taskHandler() db insert error (%s)\n", err)
				http.Error(w, "Failed to create Task.", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			bs, _ := json.MarshalIndent(t, "", "\t")
			w.Write(bs)
		case "PUT":
			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				log.Printf("taskHandler() json decoding error:\n%s\n", err)
				http.Error(w, "Invalid task.", http.StatusBadRequest)
				return
			}

			err = updateTask(db, &t)
			if err != nil {
				log.Printf("taskHandler() db update error (%s)\n", err)
				http.Error(w, "Failed to update Task.", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case "GET":
			// /task/
			sre := `^/task/?$`
			re := regexp.MustCompile(sre)
			if re.MatchString(r.URL.Path) {
				tt, err := selectTask(db, "")
				if err != nil {
					log.Printf("taskHandler() db select error (%s)\n", err)
					http.Error(w, "Failed to query Tasks.", http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				bs, _ := json.MarshalIndent(tt, "", "\t")
				w.Write(bs)
				return
			}

			// /task/[n]
			// Ex. /task/123
			sre = `^/task(?:/(\d+)/?)?$`
			re = regexp.MustCompile(sre)
			matches := re.FindStringSubmatch(r.URL.Path)
			if matches != nil {
				id := matches[1]
				nid, _ := strconv.ParseInt(id, 10, 64)
				t, err := selectOneTask(db, nid)
				if err != nil {
					log.Printf("taskHandler() db select error (%s)\n", err)
					http.Error(w, "Failed to query Task.", http.StatusInternalServerError)
					return
				}

				if t == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				bs, _ := json.MarshalIndent(t, "", "\t")
				w.Write(bs)
				return
			}

			// Something other than /task/ or /task/nnn was requested.
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
