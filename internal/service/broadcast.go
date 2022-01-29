package service

import (
	"sync"

	"gbu-queue-api/internal/entity"
)

// broadcast allows to send message in multiple channels called listeners.
type broadcast struct {
	// It is the time when someone could say that generics are cool.
	// But they are still not. I don't want them to appear in Go.
	listeners map[chan<- entity.Post]struct{}
	mu        *sync.Mutex
}

// subscribe adds listener to a broadcast.
func (b *broadcast) subscribe() chan entity.Post {
	b.mu.Lock()
	defer b.mu.Unlock()

	listener := make(chan entity.Post)
	b.listeners[listener] = struct{}{}
	return listener
}

// unsubscribe removes listener from broadcast.
// Channel returned with subscribe() method should be providen.
func (b *broadcast) unsubscribe(listener chan entity.Post) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.listeners, listener)
}

// publish sends post to all registered listeners in broadcast.
func (b *broadcast) publish(post entity.Post) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for l := range b.listeners {
		l <- post
	}
}

func (b *broadcast) listenersCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.listeners)
}
