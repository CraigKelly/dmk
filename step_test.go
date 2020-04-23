package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

//Important: these step tests are actually using function in main.go

func TestSlowBuild(t *testing.T) {
	assert := assert.New(t)

	var err error

	log.SetFlags(0)
	assert.NoError(os.Chdir("./res"))
	defer func() {
		assert.NoError(os.Chdir(".."))
	}()

	cfgText, err := ioutil.ReadFile("slowbuild.yaml")
	assert.NoError(err)
	assert.NotEmpty(cfgText)

	cfg, err := ReadConfig(cfgText)
	assert.NoError(err)
	assert.Len(cfg, 3) //sanity check

	var missing bool
	verb := log.New(ioutil.Discard, "", 0) //log.New(os.Stdout, "", 0)

	// Clean and test files
	assert.Equal(0, DoClean(cfg, verb))
	missing, err = AnyMissing([]string{"file1.txt", "file2.txt", "combined.txt"})
	assert.NoError(err)
	assert.True(missing)

	assert.Equal(0, DoBuild(cfg, verb))
	missing, err = AnyMissing([]string{"file1.txt", "file2.txt", "combined.txt"})
	assert.NoError(err)
	assert.False(missing)

	assert.Equal(0, DoBuild(cfg, verb))
	missing, err = AnyMissing([]string{"file1.txt", "file2.txt", "combined.txt"})
	assert.NoError(err)
	assert.False(missing)

	// Clean and test files
	assert.Equal(0, DoClean(cfg, verb))
	missing, err = AnyMissing([]string{"file1.txt", "file2.txt", "combined.txt"})
	assert.NoError(err)
	assert.True(missing)
}
