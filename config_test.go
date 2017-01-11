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
	assert.Equal("xform i{1,2,3}.txt", step1.Command)
	assert.Equal([]string{"i1.txt", "i2.txt", "i3.txt"}, step1.Inputs)
	assert.Equal([]string{"o1.txt", "o2.txt", "o3.txt"}, step1.Outputs)
	assert.Equal([]string{"a.aux", "b.log"}, step1.Clean)

	step2 := cfg["step2"]
	assert.Equal("step2", step2.Name)
	assert.Equal("cmd", step2.Command)
	assert.Equal([]string{"test.txt"}, step2.Inputs)
	assert.Equal([]string{"output.bin"}, step2.Outputs)
	assert.Len(step2.Clean, 0)

	depstep := cfg["depstep"]
	assert.Equal("depstep", depstep.Name)
	assert.Equal("cmd2", depstep.Command)
	assert.Equal([]string{"o3.txt", "output.bin"}, depstep.Inputs)
	assert.Equal([]string{"combination.output"}, depstep.Outputs)
	assert.Equal([]string{"need-cleaning.1", "need-cleaning.2", "need-cleaning.3"}, depstep.Clean)
}
