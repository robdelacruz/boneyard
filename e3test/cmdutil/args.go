package cmdutil

import (
	"strconv"
	"unicode"
)

func ConvInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return n, true
}

func RemoveDups(args []string) []string {
	var retArgs []string
	argsMap := map[string]bool{}

	for _, arg := range args {
		_, exists := argsMap[arg]
		if !exists {
			retArgs = append(retArgs, arg)

			argsMap[arg] = true
		}

	}

	return retArgs
}

func FlagOn(options map[string]string, flag string) bool {
	_, ok := options[flag]
	return ok
}

// Extract args from string cmd.
// Handles double/single quotes, escaped chars, like bash.
func ParseToShellArgs(cmd string) []string {
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
