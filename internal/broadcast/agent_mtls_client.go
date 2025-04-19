package broadcast

import (
	"sync"

	"google.golang.org/grpc"
)

// AgentMTLSClientUpdate carries the new gRPC client connection when certs rotate
type AgentMTLSClientUpdate struct {
	ClientConn *grpc.ClientConn
}

// AgentMTLSClientBroadcaster broadcasts updated mTLS Agent gRPC client connection
type AgentMTLSClientBroadcaster struct {
	mu   sync.RWMutex
	subs []chan AgentMTLSClientUpdate
	last *grpc.ClientConn
}

// NewAgentMTLSClientBroadcaster returns a new broadcaster instance
func NewAgentMTLSClientBroadcaster() *AgentMTLSClientBroadcaster {
	return &AgentMTLSClientBroadcaster{
		subs: make([]chan AgentMTLSClientUpdate, 0),
	}
}

// Subscribe allows subscribers to receive the latest Agent client connection
func (cb *AgentMTLSClientBroadcaster) Subscribe() <-chan AgentMTLSClientUpdate {
	ch := make(chan AgentMTLSClientUpdate, 1)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.subs = append(cb.subs, ch)

	if cb.last != nil {
		ch <- AgentMTLSClientUpdate{ClientConn: cb.last}
	}

	return ch
}

// Publish sends the new Agent client connection to all subscribers,
// and closes the previous one if present
func (cb *AgentMTLSClientBroadcaster) Publish(conn *grpc.ClientConn) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.last != nil {
		_ = cb.last.Close()
	}

	cb.last = conn

	for _, sub := range cb.subs {
		select {
		case sub <- AgentMTLSClientUpdate{ClientConn: conn}:
		default:
		}
	}
}
