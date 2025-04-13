package broadcast

import (
	"net/http"
	"sync"
)

// MTLSClientUpdate is the updated mTLS client sent over channels to subscribers
type MTLSClientUpdate struct {
	Client *http.Client
}

// MTLSClientBroadcaster is the pub-sub mechanism for broadcasting the newest mTLS client (with new certs)
// to all the managers that depend on an mTLS client
type MTLSClientBroadcaster struct {
	mu   sync.RWMutex
	subs []chan MTLSClientUpdate
	last *http.Client
}

// NewMTLSClientBroadcaster returns an instance of the MTLSClientBroadcaster
func NewMTLSClientBroadcaster() *MTLSClientBroadcaster {
	return &MTLSClientBroadcaster{
		subs: make([]chan MTLSClientUpdate, 0),
	}
}

// Subscribe is the method subscribers call to receive the latest mTLS client
func (cb *MTLSClientBroadcaster) Subscribe() <-chan MTLSClientUpdate {
	ch := make(chan MTLSClientUpdate, 1)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.subs = append(cb.subs, ch)
	if cb.last != nil {
		ch <- MTLSClientUpdate{Client: cb.last}
	}
	return ch
}

// Publish is the method publishers call to publish the latest mTLS client
func (cb *MTLSClientBroadcaster) Publish(client *http.Client) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.last = client
	for _, sub := range cb.subs {
		select {
		case sub <- MTLSClientUpdate{Client: client}:
		default:
		}
	}
}
