package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastMultiListener(t *testing.T) {
	assert := assert.New(t)

	b := NewBroadcaster()
	pcheck(b.Start())

	wg := sync.WaitGroup{}
	started := make(chan bool)
	counts := make(chan int, 8)
	stdRoutines := 1

	for i := 0; i < stdRoutines; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()

			count := 0
			t.Logf("LIST %d started\n", num)
			li := b.GetListener()
			started <- true

			for msg := range li.Delivery {
				t.Logf("LIST %d: (%v) %s\n", num, msg.CreateTime, msg.Msg)
				li.Respond(true)
				count++
			}

			counts <- count
		}(i)

		<-started
	}

	pcheck(b.Send("Message A"))
	pcheck(b.Send("Message B"))
	pcheck(b.Send("Message C"))

	diffCount := make(chan int)
	wg.Add(1)
	go func() {
		wg.Done()

		count := 0
		t.Logf("LIST X started\n")
		li := b.GetListener()
		started <- true

		for msg := range li.Delivery {
			t.Logf("LIST X: (%v) %s\n", msg.CreateTime, msg.Msg)
			count++
			if count >= 4 {
				li.Respond(false)
				break
			}
			li.Respond(true)
		}

		diffCount <- count
	}()
	<-started

	pcheck(b.Send("Message D [last for X]"))
	pcheck(b.Send("Message LAST"))

	b.Kill()
	wg.Wait()

	for i := 0; i < stdRoutines; i++ {
		assert.Equal(5, <-counts)
	}
	assert.Equal(4, <-diffCount)
}
