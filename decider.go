package main

import (
	"errors"
	"os"
	"time"
)

// Decider is something that determines if a build step should run
type Decider interface {
	// NeedBuild returns true if the build should continue
	NeedBuild(inputs []string, outputs []string) (bool, error)
}

// TimeDecider forces a build if any input is newer than any output
// This is the default build decider
type TimeDecider struct{}

// NeedBuild - return true if need a build
func (td TimeDecider) NeedBuild(inputs []string, outputs []string) (bool, error) {
	if len(outputs) < 1 {
		return false, errors.New("Nothing to build")
	}

	if missing, err := anyMissing(outputs); missing || err != nil {
		// Either we have an output missing or an error: either way we're done
		return missing, err
	}

	inputMaxTime, err := maxTime(inputs)
	if err != nil {
		return false, err
	}
	outputMinTime, err := minTime(outputs)
	if err != nil {
		return false, err
	}
	if outputMinTime.Before(inputMaxTime) {
		return true, nil // Need a build
	}
	return false, nil // Everything OK - no build
}

func maxTime(files []string) (time.Time, error) {
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

func minTime(files []string) (time.Time, error) {
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

// Note that this is returned by NeedBuild above
func anyMissing(files []string) (bool, error) {
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
