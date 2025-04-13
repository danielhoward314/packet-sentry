package broadcast

import (
	"sync"
)

const (
	// CommandGetBPFConfig tells the pcap manager to fetch BPF config from the server
	CommandGetBPFConfig = "get_bpf_config"
	// CommandSendInterfaces tells the pcap manager to send all interfaces available for capture
	CommandSendInterfaces = "send_interfaces"
)

// Command is the command sent over channels to subscribers
type Command struct {
	Name string `json:"name"`
}

// CommandsBroadcaster is the pub-sub mechanism for broadcasting commands
type CommandsBroadcaster struct {
	mu   sync.RWMutex
	subs []chan Command
	last *Command
}

// NewCommandsBroadcaster returns an instance of the CommandsBroadcaster
func NewCommandsBroadcaster() *CommandsBroadcaster {
	return &CommandsBroadcaster{
		subs: make([]chan Command, 0),
	}
}

// Subscribe is the method subscribers call to receive the latest command
func (cb *CommandsBroadcaster) Subscribe() <-chan Command {
	ch := make(chan Command, 1)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.subs = append(cb.subs, ch)
	if cb.last != nil {
		ch <- Command{Name: cb.last.Name}
	}
	return ch
}

// Publish is the method publishers call to publish the latest command received
func (cb *CommandsBroadcaster) Publish(command *Command) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.last = command
	for _, sub := range cb.subs {
		select {
		case sub <- Command{Name: command.Name}:
		default:
		}
	}
}
