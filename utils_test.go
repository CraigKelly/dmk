package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

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

func TestTimeChecks(t *testing.T) {
	assert := assert.New(t)

	var ti time.Time
	var e error

	// Some super simple testing
	ti, e = MinTime([]string{})
	assert.Equal(time.Time{}, ti)
	assert.NoError(e)

	ti, e = MaxTime([]string{})
	assert.Equal(time.Time{}, ti)
	assert.NoError(e)

	// Fail on missing file
	ti, e = MinTime([]string{"/nothing/to/read"})
	assert.Equal(time.Time{}, ti)
	assert.Error(e)

	ti, e = MaxTime([]string{"/nothing/to/read"})
	assert.Equal(time.Time{}, ti)
	assert.Error(e)

	// Need a test file and it's time stamp
	tmp, e := ioutil.TempFile("", "dmktest")
	assert.Nil(e)
	assert.NotNil(tmp)
	defer os.Remove(tmp.Name())
	if _, we := tmp.Write([]byte("yadda")); we != nil {
		assert.Fail(we.Error())
	}
	tmp.Close()
	st, e := os.Stat(tmp.Name())
	expect := st.ModTime()

	// Single file testing
	ti, e = MinTime([]string{tmp.Name()})
	assert.Equal(expect, ti)
	assert.NoError(e)

	ti, e = MaxTime([]string{tmp.Name()})
	assert.Equal(expect, ti)
	assert.NoError(e)

	// Multi file testing
	ti, e = MinTime([]string{tmp.Name(), tmp.Name(), tmp.Name()})
	assert.Equal(expect, ti)
	assert.NoError(e)

	ti, e = MaxTime([]string{tmp.Name(), tmp.Name(), tmp.Name()})
	assert.Equal(expect, ti)
	assert.NoError(e)

	ti, e = MinTime([]string{tmp.Name(), tmp.Name(), "/dev/null"})
	// assert.Equal(expect, ti)
	assert.NoError(e)

	ti, e = MaxTime([]string{"/dev/null", tmp.Name(), tmp.Name()})
	assert.Equal(expect, ti)
	assert.NoError(e)

	// Multi file with error
	ti, e = MinTime([]string{tmp.Name(), tmp.Name(), "/nothing/to/read"})
	assert.Equal(time.Time{}, ti)
	assert.Error(e)

	ti, e = MaxTime([]string{tmp.Name(), tmp.Name(), "/nothing/to/read"})
	assert.Equal(time.Time{}, ti)
	assert.Error(e)
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

	// Final test - make sure we correctly handle bad patterns
	found, err = MultiGlob([]string{"a", "util*.go", "[]a]"})
	assert.Equal([]string{}, found)
	assert.Error(err)
}
