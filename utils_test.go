package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueStrings(t *testing.T) {
	assert := assert.New(t)

	us := NewUniqueStrings()
	assert.Len(us.strings, 0)

	us.Add("b")
	us.Add("a")
	assert.Len(us.strings, 2)

	us.Add("b")
	us.Add("a")
	assert.Len(us.strings, 2)

	assert.Equal([]string{"a", "b"}, us.Strings())
	assert.Equal([]string{"a", "b"}, us.Strings())
}
