package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// TODO: add readme docs for abstract/base stuff
// TODO: add readme docs for vars section
// TODO: add readme docs for -listSteps and bash completion
// TODO: make sure docs are clean about the DMK_* env vars are for
//       COMMANDS RUNNING, not STEPS (although DMK_STEPNAME is avail)
// TODO: add tests for abstract and vars functionality
// TODO: after above 3 items we have a new release

// ConfigFile represents all the data read from a config file
type ConfigFile map[string]*BuildStep

// BuildStep is a single step in a ConfigFile
type BuildStep struct {
	Name      string            // Set after parsing (not in config file)
	Command   string            `yaml:"command"`
	Inputs    []string          `yaml:"inputs"`
	Outputs   []string          `yaml:"outputs"`
	Clean     []string          `yaml:"clean"`
	Explicit  bool              `yaml:"explicit"`
	DelOnFail bool              `yaml:"delOnFail"`
	Direct    bool              `yaml:"direct"`
	Abstract  bool              `yaml:"abstract"`
	BaseStep  string            `yaml:"baseStep"`
	Vars      map[string]string `yaml:"vars"`
}

// ReadConfig parses and returns the contents of the config file (or an error)
func ReadConfig(fileContent []byte) (ConfigFile, error) {
	// Parse the YAML
	cfg := ConfigFile{}
	err := yaml.Unmarshal(fileContent, cfg)
	if err != nil {
		return nil, err
	}

	cfg, abstractCfg, err := splitAbstractSteps(cfg)
	if err != nil {
		return nil, err
	}

	// Perform post-parse-processing (the dreaded triple P!)
	for name, step := range cfg {
		// Manually set build step name
		step.Name = name

		// If they didn't supply a map then we need to create one
		if step.Vars == nil {
			step.Vars = make(map[string]string)
		}

		// Trim any whitespace from the command so they can use YAML multi-line
		step.Command = strings.TrimSpace(step.Command)

		// If this step has a base step, grab it's data
		if len(step.BaseStep) > 0 {
			abs, absok := abstractCfg[step.BaseStep]
			if !absok {
				return nil, errors.New("No abstract step named " + step.BaseStep)
			}

			// ONLY copy command if we don't already have one
			if len(step.Command) < 1 {
				step.Command = abs.Command
			}

			// Copy properties that override
			step.Explicit = abs.Explicit
			step.DelOnFail = abs.DelOnFail
			step.Direct = abs.Direct

			// Append properties that just update
			step.Inputs = append(step.Inputs, abs.Inputs...)
			step.Outputs = append(step.Outputs, abs.Outputs...)
			step.Clean = append(step.Clean, abs.Clean...)

			// The Vars maps are different: the base map only supplies missing values
			for k, v := range abs.Vars {
				if _, ok := step.Vars[k]; !ok {
					step.Vars[k] = v
				}
			}
		}

		// Special: we add DMK_STEPNAME to the variables
		step.Vars["DMK_STEPNAME"] = step.Name

		// We allow globbing for inputs and clean
		if i, e := MultiGlob(step.Inputs); e == nil {
			step.Inputs = i
		} else {
			return nil, e
		}
		if c, e := MultiGlob(step.Clean); e == nil {
			step.Clean = c
		} else {
			return nil, e
		}

		// Expand any environment variables in command and input/clean/output
		mapping := func(envKey string) string {
			if val, ok := step.Vars[envKey]; ok {
				return val
			}
			return os.Getenv(envKey)
		}

		step.Command = os.Expand(step.Command, mapping)
		for i, t := range step.Inputs {
			step.Inputs[i] = os.Expand(t, mapping)
		}
		for i, t := range step.Outputs {
			step.Outputs[i] = os.Expand(t, mapping)
		}
		for i, t := range step.Clean {
			step.Clean[i] = os.Expand(t, mapping)
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

// splitAbstractSteps returns two config files: the main config with all
// abstract steps removed and another with only the abstract steps
func splitAbstractSteps(cfg ConfigFile) (norm ConfigFile, abstract ConfigFile, err error) {
	norm = ConfigFile{}
	abstract = ConfigFile{}
	for name, step := range cfg {
		if step.Abstract {
			abstract[name] = cfg[name]
		} else {
			norm[name] = cfg[name]
		}
	}
	return
}
