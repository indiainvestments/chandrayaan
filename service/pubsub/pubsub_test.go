package pubsub

import (
	"fmt"
	"sync"
	"testing"
)

func TestPubSub(t *testing.T) {

	c1 := make(chan interface{})
	c2 := make(chan interface{})
	c3 := make(chan interface{})

	pubsub := NewPubSub()
	pubsub.Start()

	pubsub.Register(c1)
	pubsub.Register(c2)
	pubsub.Register(c3)

	pubsub.Broadcast(42)

	wg := sync.WaitGroup{}

	wg.Add(3)

	go func() {
		defer wg.Done()
		fmt.Println(<-c1)
	}()

	go func() {
		defer wg.Done()
		fmt.Println(<-c2)
	}()

	go func() {
		defer wg.Done()
		fmt.Println(<-c3)
	}()

	pubsub.Stop()

	wg.Wait()

}
