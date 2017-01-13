package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueStrings(t *testing.T) {
	assert := assert.New(t)

	us := NewUniqueStrings()
	assert.Len(us.Seen, 0)

	us.Add("b")
	us.Add("a")
	assert.Len(us.Seen, 2)

	us.Add("b")
	us.Add("a")
	assert.Len(us.Seen, 2)

	// Test twice
	assert.Equal([]string{"a", "b"}, us.Strings())
	assert.Equal([]string{"a", "b"}, us.Strings())
}

func TestFindFirstFile(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("utils_test.go", FirstFileFound("utils_test.go", "not-here"))
	assert.Equal("utils_test.go", FirstFileFound("not-here", "utils_test.go"))
	assert.Equal("utils_test.go", FirstFileFound("not-here", "utils_test.go", "not-here"))

	assert.Equal("", FirstFileFound())
	assert.Equal("", FirstFileFound(""))
	assert.Equal("", FirstFileFound("."))
	assert.Equal("", FirstFileFound(".."))
}

func TestMultiGlob(t *testing.T) {
	assert := assert.New(t)

	var found []string
	var err error

	found, err = MultiGlob([]string{})
	assert.Equal([]string{}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{""})
	assert.Equal([]string{}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"", " ", "a", "", "\t"})
	assert.Equal([]string{"a"}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"a"})
	assert.Equal([]string{"a"}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"a", "b"})
	assert.Equal([]string{"a", "b"}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"util*.go"})
	assert.Equal([]string{"utils.go", "utils_test.go"}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"a", "util*.go", "b"})
	assert.Equal([]string{"a", "b", "utils.go", "utils_test.go"}, found)
	assert.NoError(err)

	found, err = MultiGlob([]string{"utils.go", "util*.go"})
	assert.Equal([]string{"utils.go", "utils_test.go"}, found)
	assert.NoError(err)
}
