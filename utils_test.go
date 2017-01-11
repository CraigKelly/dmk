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
