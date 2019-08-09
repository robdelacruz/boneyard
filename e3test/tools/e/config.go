package main

import (
	"bufio"
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
)

func readOptions() (map[string]string, map[string]string, []string, error) {
	options := map[string]string{}
	aliases := map[string]string{}

	fconfig := flag.String("conf", "", "config file")
	flogfile := flag.String("logfile", "", "log filename")
	fverbose := flag.Bool("verbose", false, "output info messages")

	fhttp := flag.Bool("http", false, "http server")
	feval := flag.String("e", "", "run command")

	flag.Parse()

	// -conf configfile
	err := loadConfig(*fconfig, options, aliases)
	if err != nil {
		return options, aliases, flag.Args(), err
	}

	// Command-line flags
	if *flogfile != "" {
		options["logfile"] = *flogfile
	}
	if *fverbose {
		options["verbose"] = ""
	}
	if *fhttp {
		options["http"] = ""
	}
	options["eval"] = *feval

	return options, aliases, flag.Args(), nil
}

func loadConfig(configFile string, options, aliases map[string]string) error {
	// If config file not specified, check other dirs in this order:
	// ~/.e3conf
	// /etc/e3.conf
	if configFile == "" {
		user, err := user.Current()
		if err == nil {
			configFile = filepath.Join(user.HomeDir, ".e3conf")
		} else {
			configFile = "/etc/e3.conf"
		}
	}
	confSettings, confAliases, err := readConfig(configFile)
	if err != nil {
		return err
	}

	for k, v := range confSettings {
		options[k] = v
	}
	for k, v := range confAliases {
		aliases[k] = v
	}
	return nil
}

// settings, aliases, err := readConfig("./.e3conf")
func readConfig(file string) (map[string]string, map[string]string, error) {
	settings := map[string]string{}
	aliases := map[string]string{}

	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		parseConfigLine(l, settings, aliases)
	}
	err = scanner.Err()
	if err != nil {
		return nil, nil, err
	}

	return settings, aliases, nil
}

func parseConfigLine(l string, settings, aliases map[string]string) {
	var k, setk, v string
	var fAlias bool

	for _, c := range l {
		// Skip leading whitespace
		if k == "" && unicode.IsSpace(c) {
			continue
		}

		// key=
		if setk == "" && c == '=' {
			k = strings.TrimSpace(k)
			if k == "" {
				return
			}
			setk = k
			continue
		}

		// key=...
		if setk != "" {
			v += string(c)
			continue
		}

		// 'alias '
		if setk == "" && unicode.IsSpace(c) && !fAlias && k == "alias" {
			fAlias = true
			k = ""
			continue
		}

		// key...
		k += string(c)
	}

	setk = strings.TrimSpace(setk)
	v = strings.TrimSpace(v)
	if setk == "" || v == "" {
		return
	}

	// Strip out any surrounding quotes: "val" or 'val' => val
	if strings.HasPrefix(v, "\"") {
		v = strings.Trim(v, "\"")
	}
	if strings.HasPrefix(v, "'") {
		v = strings.Trim(v, "'")
	}

	mapSet := settings
	if fAlias {
		mapSet = aliases
	}
	mapSet[setk] = v
}
