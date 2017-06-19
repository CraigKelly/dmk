package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileRead(t *testing.T) {
	assert := assert.New(t)

	cfgText, err := ioutil.ReadFile("res/test.Pipeline")
	pcheck(err)

	assert.NotEmpty(cfgText)

	cfg, err := ReadConfig(cfgText)
	pcheck(err)

	assert.NotEmpty(cfg)
	assert.Len(cfg, 3)
	assert.Contains(cfg, "step1")
	assert.Contains(cfg, "step2")
	assert.Contains(cfg, "depstep")

	step1 := cfg["step1"]
	assert.Equal("step1", step1.Name)
	assert.Equal("xformxyz i{1,2,3}.txt", step1.Command)
	assert.Equal([]string{"i1.txt", "i2.txt", "i3.txt"}, step1.Inputs)
	assert.Equal([]string{"o1.txt", "o2.txt", "o3.txt"}, step1.Outputs)
	assert.Equal([]string{"a.aux", "b.log"}, step1.Clean)

	step2 := cfg["step2"]
	assert.Equal("step2", step2.Name)
	assert.Equal("cmd1xyz", step2.Command)
	assert.Equal([]string{"test.txt"}, step2.Inputs)
	assert.Equal([]string{"output.bin"}, step2.Outputs)
	assert.Len(step2.Clean, 0)

	depstep := cfg["depstep"]
	assert.Equal("depstep", depstep.Name)
	assert.Equal("cmd2xyz", depstep.Command)
	assert.Equal([]string{"o3.txt", "output.bin"}, depstep.Inputs)
	assert.Equal([]string{"combination.output"}, depstep.Outputs)
	assert.Equal([]string{"need-cleaning.1", "need-cleaning.2", "need-cleaning.3"}, depstep.Clean)
}

func assertSteps(assert *assert.Assertions, cfg ConfigFile, req []string, steps ...string) {
	newCfg, err := TrimSteps(cfg, req)
	assert.NoError(err)

	gather := NewUniqueStrings()
	for s := range newCfg {
		gather.Add(s)
	}
	actual := gather.Strings()

	expected := make([]string, 0, len(actual)+1)
	expected = append(expected, steps...)

	assert.Equal(expected, actual)
}

func TestTrimConfig(t *testing.T) {
	assert := assert.New(t)

	cfgText, err := ioutil.ReadFile("res/trimming.yaml")
	pcheck(err)
	cfg, err := ReadConfig(cfgText)
	pcheck(err)
	assert.NotEmpty(cfg)
	assert.Len(cfg, 8)

	allSteps := []string{
		"disconnected",
		"patha1", "patha2a", "patha2b", "patha3a",
		"pathb1", "pathb2", "pathb3",
	}

	// NOP trimming should succeed
	assertSteps(assert, cfg, allSteps, allSteps...)
	assertSteps(assert, cfg, []string{}, []string{}...)

	// Error because step is missing
	_, err = TrimSteps(cfg, []string{"step-not-there"})
	assert.Error(err)

	// Check easiest possible - disconnected step
	assertSteps(assert, cfg, []string{"disconnected"}, "disconnected")

	// Single steps
	assertSteps(assert, cfg, []string{"patha1"}, "patha1")
	assertSteps(assert, cfg, []string{"pathb1"}, "pathb1")

	// get all steps from min leaves
	assertSteps(assert, cfg, []string{"disconnected", "patha3a", "pathb3"}, allSteps...)

	// Diamond dep graph
	assertSteps(assert, cfg, []string{"patha3a"}, "patha1", "patha2a", "patha2b", "patha3a")

	// Straight line dep graph
	assertSteps(assert, cfg, []string{"pathb3"}, "pathb1", "pathb2", "pathb3")
}

func TestExplicitOnly(t *testing.T) {
	assert := assert.New(t)

	var err error

	cfgText, err := ioutil.ReadFile("res/test.Pipeline")
	pcheck(err)
	cfg, err := ReadConfig(cfgText)
	pcheck(err)
	assert.Len(cfg, 3)

	var newCfg ConfigFile

	newCfg, err = NoExplicit(cfg)
	assert.NoError(err)
	assert.Len(newCfg, 3)

	cfg["depstep"].Explicit = true

	newCfg, err = NoExplicit(cfg)
	assert.NoError(err)
	assert.Len(newCfg, 2)
	newCfg, err = NoExplicit(newCfg)
	assert.NoError(err)
	assert.Len(newCfg, 2)

	newCfg["step1"].Explicit = true
	newCfg["step2"].Explicit = true

	newCfg, err = NoExplicit(cfg)
	assert.NoError(err)
	assert.Len(newCfg, 0)
	newCfg, err = NoExplicit(newCfg)
	assert.NoError(err)
	assert.Len(newCfg, 0)
}
