package main

import "time"

// TODO: move to sub-package and then give better names

// BroadcastMsg is WHAT is broadcast
type BroadcastMsg struct {
	CreateTime time.Time
	Msg        string
}

// BroadcastListener is used by the Broadcaster to track listeners
type BroadcastListener struct {
	lastMsg  int
	Delivery chan BroadcastMsg
	response chan bool
}

// Broadcaster is a concurrent-safe broadcast q
type Broadcaster struct {
	incoming     chan BroadcastMsg
	newListener  chan *BroadcastListener
	quit         chan int
	shutdown     chan int
	destinations []*BroadcastListener // Everyone who needs a message
	msgs         []BroadcastMsg       // All messages that need to be sent
}

// NewBroadcaster returns a Broacaster ready to go (when someone calls Start)
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		incoming:     make(chan BroadcastMsg),
		newListener:  make(chan *BroadcastListener),
		quit:         make(chan int),
		shutdown:     make(chan int),
		destinations: make([]*BroadcastListener, 0, 8),
		msgs:         make([]BroadcastMsg, 0, 32),
	}
}

// Start the broadcast process
func (b *Broadcaster) Start() error {
	// Our actual logic to broadcast all messages to all listeners
	handleMsgs := func() {
		for _, listener := range b.destinations {
			if listener.Delivery == nil {
				continue // listener is dead
			}
			if len(b.msgs) < 1 || listener.lastMsg >= len(b.msgs)-1 {
				continue // Nothing to deliver for listener
			}
			startIdx := listener.lastMsg + 1
			for i, msg := range b.msgs[startIdx:] {
				listener.Delivery <- msg
				listener.lastMsg = startIdx + i
				if !<-listener.response {
					// Listener wants to quit
					close(listener.Delivery)
					listener.Delivery = nil
					break
				}
			}
		}
	}

	// The loop for broadcasting
	go func() {
		// Clean up all channels
		defer func() {
			for _, listener := range b.destinations {
				if listener.Delivery != nil {
					close(listener.Delivery)
					listener.Delivery = nil
				}
			}
		}()
		defer close(b.incoming)
		defer close(b.newListener)
		defer close(b.shutdown) // This is how they known we're done

		// Quit, receive a new listener, or receive a new msg
		for {
			select {
			case msg := <-b.incoming:
				b.msgs = append(b.msgs, msg)
				handleMsgs()
			case listener := <-b.newListener:
				b.destinations = append(b.destinations, listener)
				handleMsgs()
			case <-b.quit:
				return
			}
		}
	}()

	return nil
}

// Kill the broadcast process. Once this function returns, the broadcaster is
// stopped and can NOT be reused
func (b *Broadcaster) Kill() error {
	close(b.quit)
	<-b.shutdown
	return nil
}

// GetListener returns a listener channel that receives BroadcastMsg objects
func (b *Broadcaster) GetListener() *BroadcastListener {
	list := &BroadcastListener{
		lastMsg:  -1,
		Delivery: make(chan BroadcastMsg),
		response: make(chan bool),
	}
	b.newListener <- list
	return list
}

// Send a BroadcastMsg to all listeners
func (b *Broadcaster) Send(msg string) error {
	newMsg := BroadcastMsg{
		CreateTime: time.Now().Local(),
		Msg:        msg,
	}
	b.incoming <- newMsg
	return nil
}

// Respond is called after Listen to continue receiving (or not)
func (li *BroadcastListener) Respond(continueRecv bool) {
	li.response <- continueRecv
	if !continueRecv {
		close(li.response)
	}
}
