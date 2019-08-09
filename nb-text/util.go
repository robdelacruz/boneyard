package main

import (
	"math/rand"
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

func inSlc(slc []string, s string) bool {
	for _, ss := range slc {
		if s == ss {
			return true
		}
	}
	return false
}
