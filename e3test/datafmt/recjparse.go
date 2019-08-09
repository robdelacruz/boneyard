package datafmt

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func RecjsFromString(s string) Recjs {
	var recjs Recjs
	b := &bytes.Buffer{}

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "%%") {
			recj := RecjFromString(b.String())
			recjs = append(recjs, recj)

			b = &bytes.Buffer{}
			continue
		}

		fmt.Fprintf(b, "%s\n", line)
	}

	// Add last record
	if b.String() != "" {
		recj := RecjFromString(b.String())
		recjs = append(recjs, recj)
	}

	return recjs
}

func RecjFromString(s string) *Recj {
	var fields []KVTuple

	r := regexp.MustCompile("^=+(\\w+)=+$")

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()

		matches := r.FindStringSubmatch(line)
		if matches != nil {
			mlfield := matches[1]
			fields = parseMultilineFields(mlfield, scanner, fields)
			break
		}

		toks := strings.Split(line, ": ")
		if len(toks) < 2 {
			continue
		}

		k := toks[0]
		v := strings.Replace(toks[1], "\\n", "\n", -1)
		fields = append(fields, KVTuple{k, v})
	}

	return &Recj{
		Fields: fields,
	}
}

func parseMultilineFields(firstField string, scanner *bufio.Scanner, fields []KVTuple) []KVTuple {

	curField := firstField
	r := regexp.MustCompile("^=+(\\w+)=+$")
	var b = &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		matches := r.FindStringSubmatch(line)
		if matches != nil {
			fields = append(fields, KVTuple{
				K: curField,
				V: b.String(),
			})

			curField = matches[1]
			b = &bytes.Buffer{}
			continue
		}

		fmt.Fprintf(b, "%s\n", line)
	}

	fields = append(fields, KVTuple{
		K: curField,
		V: b.String(),
	})

	return fields
}
