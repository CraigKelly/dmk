package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// ConfigFile represents all the data read from a config file
type ConfigFile map[string]*BuildStep

// BuildStep is a single step in a ConfigFile
type BuildStep struct {
	Name     string   // Set after parsing (not in config file)
	Command  string   `yaml:"command"`
	Inputs   []string `yaml:"inputs"`
	Outputs  []string `yaml:"outputs"`
	Clean    []string `yaml:"clean"`
	Explicit bool     `yaml:"explicit"`
}

// ReadConfig parses and returns the contents of the config file (or an error)
func ReadConfig(fileContent []byte) (ConfigFile, error) {
	// Parse the YAML
	cfg := ConfigFile{}
	err := yaml.Unmarshal(fileContent, cfg)
	if err != nil {
		return nil, err
	}

	// Perform post-parse-processing (the dreaded triple P!)
	for name, step := range cfg {
		// Manually set build step name
		step.Name = name
		// We allow globbing for inputs and clean
		if i, e := MultiGlob(step.Inputs); err == nil {
			step.Inputs = i
		} else {
			return nil, e
		}
		if c, e := MultiGlob(step.Clean); e == nil {
			step.Clean = c
		} else {
			return nil, e
		}
	}

	return cfg, nil
}

// TrimSteps removes all steps except the ones given and their dependencies
// via a copy-and-return (the config file passed in is unchanged)
func TrimSteps(cfg ConfigFile, reqStepNames []string) (ConfigFile, error) {
	// Find our initial steps and their dependencies
	reqSteps := make(map[string]bool)
	reqDeps := make(map[string]bool)

	for _, s := range reqStepNames {
		if _, inMap := cfg[s]; !inMap {
			return nil, fmt.Errorf("%s is not in the pipeline file", s)
		}
		reqSteps[s] = true
		for _, dep := range cfg[s].Inputs {
			reqDeps[dep] = true
		}
	}

	// Keeping adding steps our deps require until we can't add no more
	foundCount := len(reqSteps)
	for {
		// Add new deps
		for name, step := range cfg {
			if _, inMap := reqSteps[name]; inMap {
				continue // already seen this one
			}
			for _, dep := range step.Outputs {
				if _, inMap := reqDeps[dep]; inMap {
					reqSteps[name] = true // New dependency
					for _, prevDep := range step.Inputs {
						reqDeps[prevDep] = true // Get anythiung the new dep needs
					}
				}
			}
		}

		if len(reqSteps) <= foundCount {
			break // Nothing found in our iteration
		}
		foundCount = len(reqSteps)
	}

	// Now copy only the steps to keep
	newCfg := ConfigFile{}
	for name := range reqSteps {
		newCfg[name] = cfg[name]
	}

	return newCfg, nil
}

// NoExplicit returns a copy of the config file with all explicit=true steps
// removed.
func NoExplicit(cfg ConfigFile) (ConfigFile, error) {
	newCfg := ConfigFile{}
	for name, step := range cfg {
		if !step.Explicit {
			newCfg[name] = cfg[name]
		}
	}

	return newCfg, nil
}
