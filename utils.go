package main

import (
	"log"
	"sort"
)

// pcheck logs a detailed error and then panics with the same msg
func pcheck(err error) {
	if err != nil {
		log.Panicf("Fatal Error: %v\n", err)
	}
}

// UniqueStrings is a simple type around a map for a (sometimes sorted)
// list of unique strings
type UniqueStrings struct {
	Seen    map[string]bool
	Sorted  bool
	strings []string
}

// NewUniqueStrings returns a new, empty set of unique strings
func NewUniqueStrings() *UniqueStrings {
	return &UniqueStrings{
		Seen:    make(map[string]bool),
		Sorted:  false,
		strings: make([]string, 0),
	}
}

// Add a string to the unique list
func (u *UniqueStrings) Add(s string) {
	if _, inMap := u.Seen[s]; !inMap {
		u.strings = append(u.strings, s)
		u.Seen[s] = true
		u.Sorted = false
	}
}

// Strings returns the sorted array of strings
func (u *UniqueStrings) Strings() []string {
	if !u.Sorted {
		sort.Strings(u.strings)
		u.Sorted = true
	}
	return u.strings
}
