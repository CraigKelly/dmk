package main

import (
	"gopkg.in/yaml.v2"
)

//TODO: allow globbing - each entry in Inputs, Outputs, and Clean can be a glob pattern

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

	// Manually set build step name
	for name, step := range cfg {
		step.Name = name
	}

	return cfg, nil
}
