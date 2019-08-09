package datafmt

import (
	"fmt"
	"io"
	"strings"
)

// Data structures and routines for record-jar data format as described
// in TAOUP (http://www.catb.org/esr/writings/taoup/html/ch05s02.html)

// Enhancement: Multiline field values supported through this syntax:
// ===fieldname1===
// A multiline field value.
// Spanning multiple lines
// ===fieldname2===
// Another multiline field value.
//
// This produces:
// fieldname1="A multiline field value.\nSpanning multiple lines\n
// fieldname2="Another multiline field value.

// Sample format supported:
//
// Field1: Val 1
// Field2: Val 2
// Field3: Val 3
// ===MultilineField4===
// This is a value spanning multiple lines.
// Here's the second line.
// And the third.
// ===MultilineField5===
// You can have multiple multiline fields
// but they can only appear at the end of the record.
// %%
// Field1: Val 1
// Field2: Val 2
// ===Description===
// A '%%' sequence acts as a record separator.
// You can also add comments after the '%%' which
// will be ignored by the parser.
//

type KVTuple struct {
	K string
	V string
}

type Recj struct {
	Fields []KVTuple
}

type Recjs []*Recj

func NewRecj() *Recj {
	return &Recj{}
}

func (recj *Recj) AddField(k, v string) {
	recj.Fields = append(recj.Fields, KVTuple{k, v})
}

func (recj *Recj) LookupCol(k string) string {
	for _, field := range recj.Fields {
		if k == field.K {
			return field.V
		}
	}
	return ""
}

func ensureNewlineEnd(s string) string {
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}
	return s
}

func (recj *Recj) WriteString(w io.Writer) {
	var mlFields []KVTuple

	for _, field := range recj.Fields {
		// Multiline fields added last
		if strings.Contains(field.V, "\n") {
			mlFields = append(mlFields, field)
			continue
		}

		fmt.Fprintf(w, "%s: %s\n", field.K, field.V)
	}

	// Write remaining multiline fields
	for _, field := range mlFields {
		fmt.Fprintf(w, "===%s===\n", field.K)
		fmt.Fprintf(w, ensureNewlineEnd(field.V))
	}
}

func (recjs Recjs) WriteString(w io.Writer) {
	// Write each recj separated by a '%%'
	for i, recj := range recjs {
		recj.WriteString(w)
		if i < len(recjs)-1 {
			io.WriteString(w, "%%\n")
		}
	}
}

func (recj *Recj) WriteRowString(w io.Writer, cols []string, colWidths map[string]int) {
	for _, col := range cols {
		fmt.Fprintf(w, "%-[1]*[2]s ", colWidths[col], recj.LookupCol(col))
	}
}

func (recjs Recjs) WriteTableString(w io.Writer, cols []string) {
	// Maximum string length of each field in collection
	colWidths := map[string]int{}
	for _, recj := range recjs {
		for _, col := range cols {
			v := recj.LookupCol(col)
			if len(v) > colWidths[col] {
				colWidths[col] = len(v)
			}
		}
	}

	// Table heading - column names underlined
	if len(recjs) > 0 {
		for _, col := range cols {
			fmt.Fprintf(w, "%-[1]*[2]s ", colWidths[col], col)
		}
		io.WriteString(w, "\n")
		for _, col := range cols {
			fmt.Fprintf(w, "%-[1]*[2]s ", colWidths[col], strings.Repeat("-", colWidths[col]))
		}
		io.WriteString(w, "\n")
	}

	// Write each recj
	for _, recj := range recjs {
		recj.WriteRowString(w, cols, colWidths)
		io.WriteString(w, "\n")
	}
}
