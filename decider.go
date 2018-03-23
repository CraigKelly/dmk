package main

import "github.com/pkg/errors"

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

	if missing, err := AnyMissing(inputs); missing || err != nil {
		// If there was an error or we couldn't find the inputs, then we can't
		// build anything (missing deps)
		if err != nil {
			return true, errors.Wrap(err, "Error checking dependncy: cannot build")
		}
		return true, errors.New("Missing a dependency: cannot build")
	}

	if missing, err := AnyMissing(outputs); missing || err != nil {
		// Either we have an output missing or an error: either way we're done
		return missing, err
	}

	inputMaxTime, err := MaxTime(inputs)
	if err != nil {
		return false, err
	}
	outputMinTime, err := MinTime(outputs)
	if err != nil {
		return false, err
	}
	if outputMinTime.Before(inputMaxTime) {
		return true, nil // Need a build
	}
	return false, nil // Everything OK - no build
}
