# Agent

## Agent goroutine tree

```
main goroutine
├── agent Start goroutine
    ├── pcapManager StartAll goroutine loops all interfaces and associated bpfs
    │   ├── packet capture <interface-bpf> Start goroutine 1
    │   └── packet capture <interface-bpf> Start goroutine 2
    │   └── packet capture <interface-bpf> Start goroutine 3
    │   └── packet capture <interface-bpf> Start goroutine n
    ├── certificateManager Start goroutine
    ├── poller Start goroutine
```

The main goroutine blocks on receiving on a shutdown channel, which only receives if the agent startup errors or if the OS tells us to shut down. On Unix, this is done with the signals `SIGINT/SIGTERM` and on Windows, since we're running as a Windows Service, this is done by Service Control Manager sending a stop or shutdown. Either case will call the `Stop` method of the agent, which will cancel the agent goroutine context and call the `Stop/StopAll` method of each of the managers.

The agent goroutine blocks on waiting for its context to be canceled. On the happy path, it spins up the long-running goroutines for each of the managers, incrementing a wait group for each one, and stays blocked waiting for its context to be canceled. During a shutdown sequence, it executes past the blocking line and waits on each of the manager goroutines to exit.

The managers run an infinite loop in their `Start/StartAll` method, selecting on cases that signal work for their manager or that their context has been canceled. If the agent's `Stop` method is called, the manager's `Stop` method is called and its context is canceled. The agent's context is passed down to each manager during their instantiation and each manager derives a child context from it, so their contexts will get canceled if the agent's is. All of the `Stop` methods use a `sync.Once` to ensure no cleanup is harmfully duplicated.


Unlike the other managers that are just a single goroutine, the pcap manager spins up child goroutines for each interface-to-BPF association that requires a live packet capture. While the pcap manager goroutine should be running as long as the agent is running, these child goroutines may be swapped out as the user creates, updates, or deletes interface-to-BPF associations in the web console. To handle these child goroutine lifecycles, the pcap manager shares its wait group with each packet capture goroutine it spawns and the packet capture goroutine increments this wait group when spinning up. For updates, deletes, or during shutdown, the pcap manager can call the packet capture's `Stop` method. The packet capture's `Stop` method cancels its context. Since the body of the goroutine follows the similar pattern of a having a select case statement for canceled context, it exits the goroutine and a deferred cleanup function is called. The cleanup function closes the handle to the live capture and decrements the wait group counter. This final step should keep the shared wait group in a good state so the pcap manager isn't ever waiting on goroutines that have already exited.