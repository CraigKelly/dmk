package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

// MaxTime returns the maximum mod time for the files
func MaxTime(files []string) (time.Time, error) {
	if len(files) < 1 {
		return time.Time{}, nil
	}

	s, err := os.Stat(files[0])
	if err != nil {
		return time.Time{}, err
	}
	maxTime := s.ModTime()

	for _, file := range files[1:] {
		s, err := os.Stat(file)
		if err != nil {
			return time.Time{}, err
		}
		t := s.ModTime()
		if t.After(maxTime) {
			maxTime = t
		}
	}

	return maxTime, nil
}

// MinTime returns minimum mod time for the files
func MinTime(files []string) (time.Time, error) {
	if len(files) < 1 {
		return time.Time{}, nil
	}

	s, err := os.Stat(files[0])
	if err != nil {
		return time.Time{}, err
	}
	minTime := s.ModTime()

	for _, file := range files[1:] {
		s, err := os.Stat(file)
		if err != nil {
			return time.Time{}, err
		}
		t := s.ModTime()
		if t.Before(minTime) {
			minTime = t
		}
	}

	return minTime, nil
}

// AnyMissing returns true if any file does not exist
func AnyMissing(files []string) (bool, error) {
	if len(files) < 1 {
		return false, nil // no files - nothing can be missing
	}

	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return true, err
		}
	}

	return false, nil
}

// FirstFileFound returns the first file that exists
func FirstFileFound(files ...string) string {
	for _, f := range files {
		if s, err := os.Stat(f); err == nil && s != nil && !s.IsDir() {
			return f
		}
	}
	return ""
}

// MultiGlob returns an array of files matching the given patterns in sorted
// order with duplicates removed. If a pattern does not appear to be a pattern
// it is added to the returned list of strings. Whitespace-only and empty
// strings are ignored.
func MultiGlob(patterns []string) ([]string, error) {
	found := NewUniqueStrings()

	for _, p := range patterns {
		if len(strings.TrimSpace(p)) < 1 {
			continue // no whitespace only (or empty) strings
		}
		if !strings.ContainsAny(p, "*?[]") {
			found.Add(p) // Not a pattern
			continue
		}
		globbed, err := filepath.Glob(p)
		if err != nil {
			return []string{}, err // Whoops
		}
		for _, g := range globbed {
			found.Add(g)
		}
	}

	return found.Strings(), nil
}
