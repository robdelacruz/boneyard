package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Ctx struct {
	SvcUri string
	UserId int64
}

func main() {
	os.Args = os.Args[1:]
	switches, parms := parseArgs(os.Args)

	ctx := Ctx{
		SvcUri: switches["server"],
		UserId: 1,
	}

	if ctx.SvcUri == "" {
		fmt.Fprintln(os.Stderr, "Please specify server (Ex. -server localhost:8000)")
		os.Exit(1)
	}

	if !strings.HasPrefix(ctx.SvcUri, "http://") {
		ctx.SvcUri = "http://" + ctx.SvcUri
	}

	if !strings.HasSuffix(ctx.SvcUri, "/") {
		ctx.SvcUri += "/"
	}

	cmd := "help"
	if len(parms) > 0 {
		if parms[0] == "list" || parms[0] == "info" || parms[0] == "help" {
			cmd = parms[0]
			parms = parms[1:]
		}
	}

	switch cmd {
	case "list":
		resp, err := http.Get(ctx.SvcUri + "task/")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Returned server error:\n(%s)\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			bs, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Returned internal server error:\n(%s)\n", string(bs))
			os.Exit(1)
		} else if resp.StatusCode >= 400 {
			fmt.Fprintln(os.Stderr, "No Tasks")
			os.Exit(0)
		} else if resp.StatusCode >= 300 {
			fmt.Fprintln(os.Stderr, "Server moved")
			os.Exit(1)
		}

		var tt []Task
		err = json.NewDecoder(resp.Body).Decode(&tt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error returned:\n(%s)\n", err)
			os.Exit(1)
		}

		for _, t := range tt {
			fmt.Print(t)
		}
	}
}

func parseArgs(args []string) (map[string]string, []string) {
	switches := map[string]string{}
	parms := []string{}

	standaloneSwitches := []string{}
	definitionSwitches := []string{"config", "user", "pwd", "server"}
	fNoMoreSwitches := false
	curKey := ""

	for _, arg := range args {
		if fNoMoreSwitches {
			// any arg after "--" is a standalone parameter
			parms = append(parms, arg)
			continue
		}
		if arg == "--" {
			// "--" means no more switches to come
			fNoMoreSwitches = true
			continue
		}
		if curKey != "" {
			switches[curKey] = arg
			curKey = ""
			continue
		}
		if strings.HasPrefix(arg, "--") {
			k := arg[2:]
			if listContains(definitionSwitches, k) {
				// --key "val"
				curKey = k
				continue
			}
			if listContains(standaloneSwitches, k) {
				// --standalone
				switches[k] = "y"
				curKey = ""
				continue
			}
		}
		if strings.HasPrefix(arg, "-") {
			k := arg[1:]
			if listContains(definitionSwitches, k) {
				// -key "val"
				curKey = k
				continue
			}
			if listContains(standaloneSwitches, k) {
				// -standalone
				switches[k] = "y"
				curKey = ""
				continue
			}
			// -xyz standalone switches (similar to -x -y -z)
			for _, ch := range k {
				k := string(ch)
				if listContains(standaloneSwitches, k) {
					switches[k] = "y"
				}
			}
			curKey = ""
			continue
		}

		// standalone parameter
		parms = append(parms, arg)
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
