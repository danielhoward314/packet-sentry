# Agent

## Agent goroutine tree

main goroutine
├── agent Start goroutine
    ├── pcapManager StartAll loops all interfaces and associated bpfs
    │   ├── packet capture <interface-bpf> goroutine 1
    │   └── packet capture <interface-bpf> goroutine 2
    │   └── packet capture <interface-bpf> goroutine 3
    │   └── packet capture <interface-bpf> goroutine n
    ├── certificateManager Start goroutine
    ├── poller Poll goroutine

