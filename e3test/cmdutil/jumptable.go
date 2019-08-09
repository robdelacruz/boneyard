package cmdutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type Req struct {
	Opts  map[string]string
	Args  []string
	Nargs map[string]string
}

type Resp struct {
	Code   int
	Status string
	Args   []string
	Nargs  map[string]string
}

func (resp *Resp) String() string {
	respArgs := strings.Join(resp.Args, " ")

	var b bytes.Buffer
	fmt.Fprintf(&b, "(%d) %s\n", resp.Code, resp.Status)
	if len(respArgs) > 0 {
		fmt.Fprintf(&b, "> Return Args: %s\n", respArgs)
	}

	return b.String()
}

type Handler func(req *Req, r io.Reader, w io.Writer) (*Resp, error)

type JumpTbl map[string]Handler

func NewJumpTbl() JumpTbl {
	return JumpTbl{}
}

func (jt JumpTbl) Handle(cmd string, fn Handler) {
	jt[cmd] = fn
}

func (jt JumpTbl) Exec(verb string, req *Req, r io.Reader, w io.Writer) (*Resp, error) {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}

	doFunc, _ := jt[verb]
	if doFunc == nil {
		return &Resp{}, nil
	}

	return doFunc(req, r, w)
}
