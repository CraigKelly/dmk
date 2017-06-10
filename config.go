package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

/* TODO: Base steps:

UPDATED: We have abstract steps, now we need to create a POC in res/vars.yaml
and actually handle variables

   1. A step marks itself as something like abstract or template
   2. "Actual" steps can list an abstract step as a base/template and will
      receive its properties as a default
   3. Every step gets a variable section where the vars get expanded. Eval
      order will be step vars, then env vars, then globbing. Abstract steps
      don't expand anything.
   4. This means two new doc sections: variables and abstract steps.
   5. Need really good testing for this stuff.
   6. Examples for variables: the clean section for rubber/latex steps
   7. Abstract steps would be handy for steps that are very similar (like
      one step per csv file extracted from a data source)

TODO: add some tests for abstract and vars yaml files
*/

// ConfigFile represents all the data read from a config file
type ConfigFile map[string]*BuildStep

// BuildStep is a single step in a ConfigFile
type BuildStep struct {
	Name      string   // Set after parsing (not in config file)
	Command   string   `yaml:"command"`
	Inputs    []string `yaml:"inputs"`
	Outputs   []string `yaml:"outputs"`
	Clean     []string `yaml:"clean"`
	Explicit  bool     `yaml:"explicit"`
	DelOnFail bool     `yaml:"delOnFail"`
	Direct    bool     `yaml:"direct"`
	Abstract  bool     `yaml:"abstract"`
	BaseStep  string   `yaml:"baseStep"`
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

		// Trim any whitespace from the command so they can use YAML multi-line
		step.Command = strings.TrimSpace(step.Command)

		// If this step has a base step, grab it's data
		if len(step.BaseStep) > 0 {
			abs, absok := abstractCfg[step.BaseStep]
			if !absok {
				return nil, errors.New("No abstract step named " + step.BaseStep)
			}

			// Copy properties that override
			step.Explicit = abs.Explicit
			step.DelOnFail = abs.DelOnFail
			step.Direct = abs.Direct

			// Append properties that just update
			step.Inputs = append(step.Inputs, abs.Inputs...)
			step.Outputs = append(step.Outputs, abs.Outputs...)
			step.Clean = append(step.Clean, abs.Clean...)
		}

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

		// Expand any environment variables in input/clean/output
		for i, t := range step.Inputs {
			step.Inputs[i] = os.ExpandEnv(t)
		}
		for i, t := range step.Outputs {
			step.Outputs[i] = os.ExpandEnv(t)
		}
		for i, t := range step.Clean {
			step.Clean[i] = os.ExpandEnv(t)
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
