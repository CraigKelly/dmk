package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Important: not much to test: we just need to make sure that the version
// function is current and working

func TestVersion(t *testing.T) {
	assert := assert.New(t)

	verTextRaw, err := ioutil.ReadFile("./VERSION")
	assert.NoError(err)
	verText := strings.TrimSpace(string(verTextRaw))
	assert.NotEmpty(verText)

	verResp := Version()
	assert.NotEmpty(verResp)

	assert.Equal(0, strings.Index(verResp, verText))
}
