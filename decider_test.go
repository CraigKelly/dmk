package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeDecider(t *testing.T) {
	assert := assert.New(t)

	var d Decider
	d = TimeDecider{}

	var b bool
	var e error

	b, e = d.NeedBuild([]string{}, []string{})
	assert.False(b)
	assert.Error(e)

	b, e = d.NeedBuild([]string{"/nothing/to/read"}, []string{})
	assert.False(b)
	assert.Error(e)

	b, e = d.NeedBuild([]string{}, []string{"/nothing/to/read"})
	assert.True(b)
	assert.Nil(e)

	o1, e := ioutil.TempFile("", "dmktest")
	assert.Nil(e)
	assert.NotNil(o1)
	defer os.Remove(o1.Name())
	if _, we := o1.Write([]byte("yadda")); we != nil {
		assert.Fail(we.Error())
	}
	o1.Close()

	// No input - must build
	b, e = d.NeedBuild([]string{}, []string{o1.Name()})
	assert.True(b)
	assert.Nil(e)

	time.Sleep(5 * time.Millisecond) // Hacky way to make sure input is newer

	i1, e := ioutil.TempFile("", "dmktest")
	assert.Nil(e)
	assert.NotNil(i1)
	defer os.Remove(i1.Name())
	if _, we := i1.Write([]byte("yadda")); we != nil {
		assert.Fail(we.Error())
	}
	i1.Close()

	// Input newer - must build
	b, e = d.NeedBuild([]string{i1.Name()}, []string{o1.Name()})
	assert.True(b)
	assert.Nil(e)

	// Input older - must NOT build
	os.Chtimes(o1.Name(), time.Now().Local(), time.Now().Local())
	b, e = d.NeedBuild([]string{i1.Name()}, []string{o1.Name()})
	assert.False(b)
	assert.Nil(e)
}
