package store

import (
	"bytes"
)

type ErrorBag struct {
	errs []error
}

func (eb *ErrorBag) Add(err error) {
	eb.errs = append(eb.errs, err)
}

func (eb *ErrorBag) HasErrors() bool {
	return len(eb.errs) > 0
}

func (eb ErrorBag) Error() string {
	var b bytes.Buffer
	for i := len(eb.errs) - 1; i >= 0; i-- {
		err := eb.errs[i]
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return b.String()
}
