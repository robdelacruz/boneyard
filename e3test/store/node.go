package store

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
	"time"

	"e3/datafmt"
)

var _tagSep = ","

func TagSep() string {
	return _tagSep
}

func (nl *NodeList) toRecjs() datafmt.Recjs {
	var recjs datafmt.Recjs
	for _, n := range nl.Items {
		recjs = append(recjs, n.toRecj())
	}
	return recjs
}

func nodeFromRecj(recj *datafmt.Recj) *Node {
	var n Node

	for _, kv := range recj.Fields {
		switch kv.K {
		case "id":
			n.ID = kv.V
		case "alias":
			n.Alias = kv.V
		case "title":
			n.Title = kv.V
		case "assigned":
			n.Assigned = kv.V
		case "body":
			n.Body = kv.V
		case "tags":
			n.Tags = strings.Split(kv.V, _tagSep)
		case "createdt":
			n.Createdt = kv.V
		case "updatedt":
			n.Updatedt = kv.V
		}
	}

	return &n
}

func nodeListFromRecjs(recjs datafmt.Recjs) *NodeList {
	var ns []*Node

	for _, recj := range recjs {
		n := nodeFromRecj(recj)
		ns = append(ns, n)
	}

	return &NodeList{ns}
}

func (nl *NodeList) WriteRecjString(w io.Writer) {
	nl.toRecjs().WriteString(w)
}

func (nl *NodeList) WriteTableString(w io.Writer, cols []string) {
	nl.toRecjs().WriteTableString(w, cols)
}

func (n *Node) HashString() string {
	s := fmt.Sprintf("%s%s%s%s%s", n.Alias, n.Title, n.Assigned, n.Body, strings.Join(n.Tags, _tagSep))
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func NodeListFromRecjString(s string) *NodeList {
	recjs := datafmt.RecjsFromString(s)
	return nodeListFromRecjs(recjs)
}

//
// internal methods
//

func isotimestr(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func parseISOTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (n *Node) toRecj() *datafmt.Recj {
	recj := datafmt.NewRecj()

	if n.ID != "" {
		recj.AddField("id", n.ID)
	}
	recj.AddField("alias", n.Alias)
	recj.AddField("title", n.Title)
	recj.AddField("assigned", n.Assigned)
	recj.AddField("body", n.Body)
	recj.AddField("tags", fmt.Sprintf("%s", strings.Join(n.Tags, _tagSep)))
	recj.AddField("createdt", n.Createdt)
	recj.AddField("updatedt", n.Updatedt)

	return recj
}

func (n *Node) ExistsTag(tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (n *Node) AddTag(tag string) {
	tag = strings.TrimSpace(tag)

	if !n.ExistsTag(tag) {
		n.Tags = append(n.Tags, tag)
	}
}

func (n *Node) RemoveTag(tag string) {
	tag = strings.TrimSpace(tag)

	for i, t := range n.Tags {
		if t == tag {
			n.Tags = append(n.Tags[:i], n.Tags[i+1:]...)
			break
		}
	}
}

func (n *Node) ProcessCsvTags(csvTags string, fn func(tag string)) {
	tags := strings.Split(csvTags, _tagSep)
	for _, tag := range tags {
		fn(tag)
	}
}
