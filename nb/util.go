package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

const _pushchars = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

var _lastGenTime int64
var _idxs []int

func init() {
	rand.Seed(time.Now().UnixNano())
	_idxs = make([]int, 12)
}

func GenID() string {
	// ms since epoch
	now := time.Now().UnixNano() / 1000000

	if now != _lastGenTime {
		// Generate random indexes to pushcars
		for i := 0; i < 12; i++ {
			_idxs[i] = rand.Intn(64)
		}
	} else {
		// Same function call time, so use incremented previous indexes
		for i := 0; i < 12; i++ {
			_idxs[i] += 1
			if _idxs[i] < 64 {
				break
			}
			_idxs[i] = 0 // carry inc to next index pos
		}
	}
	_lastGenTime = now

	// Prepare blank 20-char ID
	idchars := make([]byte, 20)

	// Set cols 7 to 0 with time-indexed pushcars
	for i := 7; i >= 0; i-- {
		idchars[i] = _pushchars[now%64]
		now = now / 64
	}

	// Set cols 8 to 19 with randomly indexed pushchars
	iIdxs := 0
	for i := 8; i < 20; i++ {
		idx := _idxs[iIdxs]
		idchars[i] = _pushchars[idx]
		iIdxs++
	}

	return string(idchars)
}

// Return slice of map keys (keys in random order)
func MapGetKeys(mm map[string]interface{}) []string {
	kk := []string{}
	for m := range mm {
		kk = append(kk, m)
	}
	sort.Strings(kk)
	return kk
}

// Return slice of node field values corresponding to order of kk keys
func MapGetVals(node Node, kk []string) []interface{} {
	vv := make([]interface{}, 0)
	for _, k := range kk {
		vv = append(vv, node[k])
	}
	return vv
}

func SlcContains(ss []string, n string) bool {
	for _, s := range ss {
		if s == n {
			return true
		}
	}
	return false
}
func SlcRemove(ss []string, n string) []string {
	vv := []string{}
	for _, s := range ss {
		if s == n {
			continue
		}
		vv = append(vv, s)
	}
	return vv
}
func SlcInsertFirst(ss []string, n string) []string {
	vv := []string{}
	vv = append(vv, n)
	vv = append(vv, ss...)
	return vv
}

// Return elements from aa that are not in bb (aa - bb).
func SlcDiff(aa []string, bb []string) []string {
	dd := []string{}
	for _, a := range aa {
		if !SlcContains(bb, a) {
			dd = append(dd, a)
		}
	}
	return dd
}

func StrV(v interface{}) string {
	var s string
	switch n := v.(type) {
	case int:
		s = fmt.Sprintf("%d", n)
	case int32:
		s = fmt.Sprintf("%d", n)
	case int64:
		s = fmt.Sprintf("%d", n)
	case string:
		s = fmt.Sprintf("%s", n)
	case float32:
		s = fmt.Sprintf("%.2f", n)
	case float64:
		s = fmt.Sprintf("%.2f", n)
	case nil:
		s = fmt.Sprintf("")
	default:
		s = fmt.Sprintf("??")
	}
	return s
}

// Return whether jj is a subset map of ii
// (subset, meaning it all key-value pairs of ii exists in jj)
func MapContains(ii map[string]interface{}, jj map[string]interface{}) bool {
	for k, v := range jj {
		//if StrV(ii[k]) != StrV(v) {
		if ii[k] != v {
			fmt.Printf("mismatch ii[k]='%s'\nv='%s'\n", ii[k], v)
			return false
		}
	}
	return true
}
