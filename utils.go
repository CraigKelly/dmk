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
	Seen map[string]bool
}

// NewUniqueStrings returns a new, empty set of unique strings
func NewUniqueStrings() *UniqueStrings {
	return &UniqueStrings{
		Seen: make(map[string]bool),
	}
}

// Add a string to the unique list
func (u *UniqueStrings) Add(s string) {
	u.Seen[s] = true
}

// Strings returns the sorted array of strings
func (u *UniqueStrings) Strings() []string {
	strings := make([]string, 0, len(u.Seen))
	for k := range u.Seen {
		strings = append(strings, k)
	}
	sort.Strings(strings)
	return strings
}
