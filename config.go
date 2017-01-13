package main

import "gopkg.in/yaml.v2"

// ConfigFile represents all the data read from a config file
type ConfigFile map[string]*BuildStep

// BuildStep is a single step in a ConfigFile
type BuildStep struct {
	Name    string   // Set after parsing (not in config file)
	Command string   `yaml:"command"`
	Inputs  []string `yaml:"inputs"`
	Outputs []string `yaml:"outputs"`
	Clean   []string `yaml:"clean"`
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
