package core

import (
	"bytes"
	"e3/cmdutil"
	"e3/osutil"
	"e3/store"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
)

type E3C struct {
	st      *store.Store
	opts    map[string]string
	aliases map[string]string
	logger  *log.Logger
}

func NewE3C(st *store.Store, opts, aliases map[string]string, logger *log.Logger) *E3C {
	return &E3C{st, opts, aliases, logger}
}

func existsFile(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func (e3c *E3C) CreateSqliteDBFile(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	dbfile := "node.db"
	if len(req.Args) > 0 {
		dbfile = req.Args[0]
	}
	if !strings.HasSuffix(dbfile, ".db") {
		dbfile += ".db"
	}

	if existsFile(dbfile) {
		return nil, fmt.Errorf("file %s already exists in this directory", dbfile)
	}

	fmt.Fprintf(w, "Creating new database...\n")

	st := store.NewStore("sqlite3", dbfile, "", e3c.logger)
	err := st.InitTables()
	if err != nil {
		return nil, fmt.Errorf("error creating new db file %s (%s)", dbfile, err)
	}

	fmt.Fprintf(w, "New dbfile: %s\n", dbfile)
	resp := &cmdutil.Resp{
		Code:   0,
		Status: dbfile,
	}
	return resp, nil
}

// Create and initialize database schemas.
// Input request:
// Nargs["droptables"]
// createdb -droptables to drop tables before creating them.
func (e3c *E3C) Createdb(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	if cmdutil.FlagOn(req.Nargs, "droptables") {
		err := e3c.st.DropTables()
		if err != nil {
			return nil, err
		}
	}

	err := e3c.st.InitTables()
	if err != nil {
		return nil, err
	}

	return &cmdutil.Resp{}, nil
}

func pberr(err error) error {
	return fmt.Errorf("protobuf error (%s)", err)
}

func readNodeList(r io.Reader, nfmt string) (*store.NodeList, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("update: error reading from input (%s)", err)
	}

	nl := &store.NodeList{}
	if nfmt == "" || nfmt == "recj" {
		nl = store.NodeListFromRecjString(string(bs))
	} else if nfmt == "protobuf" || nfmt == "pb" {
		err := proto.Unmarshal(bs, nl)
		if err != nil {
			return nil, pberr(err)
		}
	}

	return nl, nil
}

func writeNodeList(w io.Writer, nl *store.NodeList, nfmt string) error {
	if nfmt == "" || nfmt == "recj" {
		nl.WriteRecjString(w)
	} else if nfmt == "table" {
		nl.WriteTableString(w, []string{"id", "assigned", "title", "tags"})
	} else if nfmt == "protobuf" || nfmt == "pb" {
		bs, err := proto.Marshal(nl)
		if err != nil {
			return pberr(err)
		}
		w.Write(bs)
	}

	return nil
}

// Cmd-line:
// e new {num nodes} {-field1:val} {-field2:val} ...
//
// Ex.
// e new 10                     <--- 10 new blank nodes
// e new -title="Node Title"    <--- 1 new node with title
// e new 10 -title="Node Title" <--- 10 new nodes with title
//
// Input request:
// Nargs["outputfmt"] = {recj|table|pb}
// Args = [num nodes]
// Nargs = {field1: val, field2: val, ...}
//
// Return response:
// Code = 0
// Status = message containing number of new nodes created
func (e3c *E3C) New(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	numNodes := 1

	args := req.Args
	if len(args) > 0 {
		c, ok := cmdutil.ConvInt(args[0])
		if ok {
			numNodes = c
		}
	}

	nl := store.NodeList{}
	for i := 0; i < numNodes; i++ {
		n := &store.Node{}
		n.Alias = req.Nargs["alias"]
		n.Title = req.Nargs["title"]
		n.Assigned = req.Nargs["assigned"]

		n.Body = req.Nargs["body"]
		if !strings.Contains(n.Body, "\n") {
			n.Body += "\n"
		}

		stags := req.Nargs["tags"]
		if stags != "" {
			n.Tags = strings.Split(stags, store.TagSep())
		}

		nl.Items = append(nl.Items, n)
	}

	err := writeNodeList(w, &nl, req.Nargs["outputfmt"])
	if err != nil {
		return nil, err
	}

	resp := &cmdutil.Resp{
		Code:   0,
		Status: fmt.Sprintf("%d nodes generated", numNodes),
	}
	return resp, nil
}

// Load node IDs and return recj (record-jar) text representation.
// Input request:
// Nargs["limit"] = n
// Nargs["outputfmt"] = {recj|table|pb}
// Args = list of node IDs
//
// Return response:
// sout = nodes recj text representation
// Code = number of nodes loaded
// Status = csv text list of node IDs successfully loaded
// Vals = list of node IDs successfully loaded
// Nargs["okIDs"] = list of node IDs successfully loaded
// Nargs["errIDs"] = list of node IDs failed to load
//
// Return Error: contains newline delimited error messages for each node
//               failing to load
func (e3c *E3C) Load(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	ids := cmdutil.RemoveDups(req.Args)
	if len(ids) == 0 {
		return &cmdutil.Resp{}, nil
	}

	var ns []*store.Node
	var errIDs []string
	var okIDs []string
	var err error
	var eb store.ErrorBag

	if ids[0] == "*" || ids[0] == "-" {
		nlimit, _ := cmdutil.ConvInt(req.Nargs["limit"])
		var qlimit string
		if nlimit > 0 {
			qlimit = fmt.Sprintf("limit %d", nlimit)
		}
		ns, err = e3c.st.LoadNodes("id <> ''", "id desc", qlimit)
		if err != nil {
			eb.Add(err)
		}
	} else {
		for _, id := range ids {
			n, err := e3c.st.LoadNodeByID(id)
			if err != nil {
				errIDs = append(errIDs, id)
				eb.Add(fmt.Errorf("error loading node ID %s (%s)\n", id, err))
				continue
			}

			if n != nil {
				okIDs = append(okIDs, id)
				ns = append(ns, n)
			}
		}
	}

	err = writeNodeList(w, &store.NodeList{ns}, req.Nargs["outputfmt"])
	if err != nil {
		return nil, err
	}

	resp := &cmdutil.Resp{
		Code:   len(okIDs),
		Status: strings.Join(okIDs, ","),
		Args:   okIDs,
		Nargs: map[string]string{
			"okIDs":  strings.Join(okIDs, ","),
			"errIDs": strings.Join(errIDs, ","),
		},
	}
	if eb.HasErrors() {
		return resp, eb
	}
	return resp, nil
}

// Update nodes contents.
//
// Input request:
// sin = nodes recj text representation containing updates
// Nargs["inputfmt"] = {recj|pb}
// Nargs["force"]
//
// Return response:
// Code = number of nodes successfully updated
// Status = csv text list of node IDs successfully loaded
// Vals = list of node IDs successfully updated
// Nargs["okIDs"] = list of node IDs successfully loaded
// Nargs["skippedIDs"] = list of node IDs of up to date nodes
// Nargs["errIDs"] = list of node IDs failed to load
//
// Return Error: contains newline delimited error messages for each node
//               failing to update
func (e3c *E3C) Update(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	var errIDs []string
	var skippedIDs []string
	var okIDs []string
	var berr bytes.Buffer

	nl, err := readNodeList(r, req.Nargs["inputfmt"])
	if err != nil {
		return nil, err
	}

	for _, n := range nl.Items {
		if strings.TrimSpace(n.Title) == "" {
			if n.ID != "" {
				skippedIDs = append(skippedIDs, n.ID)
			}
			e3c.logger.Printf("Node %s '%s' has no title. Skipped.\n", n.ID, n.Alias)
			continue
		}

		n.Hash = n.HashString()

		// --force bypasses the hash 'up to date' check
		if !cmdutil.FlagOn(req.Nargs, "force") {
			uptodate, _ := e3c.st.NodeIsUpToDate(n.ID, n.Hash)
			if uptodate {
				skippedIDs = append(skippedIDs, n.ID)
				e3c.logger.Printf("Node %s '%s' already up to date. Skipped.\n", n.ID, n.Alias)
				continue
			}
		}

		_, err := e3c.st.SaveNode(n)
		if err != nil {
			if n.ID != "" {
				errIDs = append(errIDs, n.ID)
			}
			fmt.Fprintf(&berr, "error saving node %s '%s' (%s)", n.ID, n.Alias, err)
		}

		okIDs = append(okIDs, n.ID)
		e3c.logger.Printf("Updated node %s '%s'\n", n.ID, n.Alias)
	}

	resp := &cmdutil.Resp{
		Code:   len(okIDs),
		Status: strings.Join(okIDs, ","),
		Args:   okIDs,
		Nargs: map[string]string{
			"okIDs":      strings.Join(okIDs, ","),
			"skippedIDs": strings.Join(skippedIDs, ","),
			"errIDs":     strings.Join(errIDs, ","),
		},
	}

	serr := berr.String()
	if len(serr) > 0 {
		return resp, errors.New(serr)
	}

	return resp, nil
}

// Search nodes and return recj text representation of nodes found.
// Input request:
// Nargs["outputfmt"] = {recj|table|pb}
// Nargs["limit"] = n
// Args[0] = search request string
//
// Return response:
// sout = found nodes recj text representation
// Code = number of nodes found
// Status = csv text list of node IDs found
// Vals = list of node IDs found
func (e3c *E3C) Search(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	if len(req.Args) == 0 && len(req.Nargs) == 0 {
		return &cmdutil.Resp{}, nil
	}

	var b bytes.Buffer

	for k, v := range req.Nargs {
		switch k {
		case "alias":
			fmt.Fprintf(&b, "Alias:%s ", v)
		case "title":
			fmt.Fprintf(&b, "Title:%s ", v)
		case "assigned":
			fmt.Fprintf(&b, "Assigned:%s ", v)
		case "body":
			fmt.Fprintf(&b, "Body:%s ", v)
		case "tags":
			fmt.Fprintf(&b, "Tags:%s ", v)
		}
	}

	for _, arg := range req.Args {
		fmt.Fprintf(&b, "%s ", arg)
	}

	q := b.String()
	ns, err := e3c.st.SearchNodes(q)
	if err != nil {
		return nil, fmt.Errorf("find error (%s)\n", err)
	}

	nlimit, _ := cmdutil.ConvInt(req.Nargs["limit"])
	if nlimit > 0 && nlimit < len(ns) {
		ns = ns[:nlimit]
	}

	err = writeNodeList(w, &store.NodeList{ns}, req.Nargs["outputfmt"])
	if err != nil {
		return nil, err
	}

	var foundIDs []string
	for _, n := range ns {
		foundIDs = append(foundIDs, n.ID)
	}
	resp := &cmdutil.Resp{
		Code:   len(ns),
		Status: strings.Join(foundIDs, ","),
		Args:   foundIDs,
	}
	return resp, nil
}

// Find nodes and return recj text representation of nodes found.
// Input request:
//
// Return response:
// sout = found nodes recj text representation
// Code = number of nodes found
// Status = csv text list of node IDs found
// Vals = list of node IDs found
func (e3c *E3C) Find(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	return &cmdutil.Resp{}, nil
}

// Reindex any new/updated nodes since the last indexing request.
func (e3c *E3C) BgIndex(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	ids, err := e3c.st.QueryChangedNodes()
	if err != nil {
		return nil, fmt.Errorf("error querying nodechange (%s)\n", err)
	}

	if len(ids) == 0 {
		fmt.Fprintf(w, "No changed nodes since last indexing.\n")
		return &cmdutil.Resp{}, nil
	}

	fmt.Fprintf(w, "Indexing %d nodes...\n", len(ids))

	for _, id := range ids {
		n, err := e3c.st.LoadNodeByID(id)
		if err != nil {
			fmt.Fprintf(w, "Skipped %s - error loading node (%s)\n", id, err)
			continue
		}

		if n == nil {
			fmt.Fprintf(w, "Skipped %s - node ID doesn't exist\n", id)

			err = e3c.st.ClearNodeChanged(id)
			if err != nil {
				fmt.Fprintf(w, "Error clearing nodechange %s (%s)\n", id, err)
			}
			continue
		}

		err = e3c.st.IndexNode(n)
		if err != nil {
			fmt.Fprintf(w, "Skipped %s %s - error indexing node (%s)\n", n.ID, n.Alias, err)
			continue
		}

		fmt.Fprintf(w, "Node %s %s added to index\n", n.ID, n.Alias)

		err = e3c.st.ClearNodeChanged(id)
		if err != nil {
			fmt.Fprintf(w, "Error clearing nodechange %s (%s)\n", id, err)
			continue
		}
	}

	return &cmdutil.Resp{}, nil
}

// Launch external editor to edit input stream.
// This copies input stream to tmp file and passes it to an external program
// such as vim to edit the input text.
// The saved text is copied to the output stream.
//
// Input request:
// sin = text to edit
//
// Return response:
// sout = saved edited text
func (e3c *E3C) Edit(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	ftmp, err := ioutil.TempFile("", "_e3")
	if err != nil {
		return nil, fmt.Errorf("error creating node tmp file (%s)", err)
	}

	// input -> tmpfile
	_, err = io.Copy(ftmp, r)
	if err != nil {
		return nil, fmt.Errorf("edit2: error writing input to tmp file %s (%s)", ftmp.Name(), err)
	}
	ftmp.Close()
	e3c.logger.Printf("Generated new node tmp file %s\n", ftmp.Name())

	err = osutil.RunCommand("gvim", "-f", ftmp.Name())
	if err != nil {
		return nil, fmt.Errorf("error running external editor (%s)", err)
	}

	ftmp, err = os.Open(ftmp.Name())
	if err != nil {
		return nil, fmt.Errorf("edit2: error opening edited tmp file (%s)", err)
	}

	_, err = io.Copy(w, ftmp)
	if err != nil {
		return nil, fmt.Errorf("edit2: error writing edits to output (%s)", err)
	}
	ftmp.Close()

	return &cmdutil.Resp{}, nil
}

// Pass input stream directly to output stream.
func (e3c *E3C) Echo(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	var b bytes.Buffer
	_, err := io.Copy(&b, r)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(os.Stdout, &b)
	if err != nil {
		return &cmdutil.Resp{}, err
	}

	_, err = io.Copy(w, &b)
	if err != nil {
		return &cmdutil.Resp{}, err
	}

	return &cmdutil.Resp{}, nil
}

// Apply operation to all input nodes.
// Used primarily to set multiple fields for each input node.
//
// map .assigned=robtwister tags+="new tag",tag2 tags-=oldtag
//   This will set field assigned to the value 'robtwister',
//   add 'new tag', and 'tag2' to node tags,
//   remove 'oldtag' from node tags (if it's present).
//
// Input request:
// sin = input nodes recj text representation
// Nargs["inputfmt"] = {recj|pb}
// Nargs["outputfmt"] = {recj|table|pb}
// Nargs[{field}] = val to assign to
//
// Return response:
// sout = updated nodes recj text representation
func (e3c *E3C) Map(req *cmdutil.Req, r io.Reader, w io.Writer) (*cmdutil.Resp, error) {
	nl, err := readNodeList(r, req.Nargs["inputfmt"])
	if err != nil {
		return nil, err
	}

	for _, n := range nl.Items {
		for k, v := range req.Nargs {
			switch k {
			case "alias":
				n.Alias = v
			case "title":
				n.Title = v
			case "title+":
				n.Title += v
			case "assigned":
				n.Assigned = v
			case "body":
				n.Body = v
			case "body+":
				n.Body += v
			case "tags+":
				n.ProcessCsvTags(v, n.AddTag)
			case "tags-":
				n.ProcessCsvTags(v, n.RemoveTag)
			}
		}
	}

	err = writeNodeList(w, nl, req.Nargs["outputfmt"])
	if err != nil {
		return nil, err
	}

	return &cmdutil.Resp{}, nil
}

func (e3c *E3C) HttpRoot(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fmt.Println("Headers:")
	fmt.Println(r.Header)

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	fmt.Println("Body:")
	fmt.Println(string(bs))
}

func (e3c *E3C) HttpCmd(w http.ResponseWriter, r *http.Request) {
	scmd, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		fmt.Println("unescape error: ", err)
		return
	}
	if scmd == "" {
		fmt.Println("No command.")
		return
	}

	fmt.Printf("Running cmd: '%s'\n", scmd)
	RunPipelineStmts(scmd, r.Body, w, e3c.st, e3c.opts, e3c.aliases, e3c.logger)
}
