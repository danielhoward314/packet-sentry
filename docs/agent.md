# Agent

The Packet Sentry Agent is the program that a user would install on endpoints to report packet capture telemetry. It runs as a daemon on Unix and as a background Service on Windows. The main Go executable in `cmd/agent/main.go` is packaged into an installer for each platform (.pkg on macOS, .msi on Windows, and .deb/.rpm on Linux).

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

## Pub-sub for mTLS agent-api gRPC unary and streaming clients

The agent communicates with the agent-api with two gRPC clients.

One client is called the bootstrap client because it bootstraps the trust between the agent and the agent-api by providing an the install key when requesting a client certificate. The agent-api validates the install key in the request and, if valid, will use the CSR in the request to issue a certificate. When the certificate is due to expire within the next 30 days, the agent will request a new one and include the fingerprint of its current certificate. The agent-api validates the fingerprint in the existing cert before issuing a new one. This bootstrap client is configured for TLS, expecting the gRPC server to present a certificate. The certificate manager is the only manager in the agent that needs this client and, since the communication is TLS, the same gRPC connection can be used over the lifetime of the agent's execution.

The other client, the agent client, invokes unary gRPCs and a streaming one. This client is only used after the agent has received its certificate, since it depends on the cert to establish a mutual TLS connection with the agent-api. Since the cert manager may renew the client certificate and the agent client is used by several managers, a pub-sub mechanism is used to notify all of the managers that the client certificate has changed. The certificate manager is the publisher and the other managers that depend on the certificate for mTLS connections are the subscribers. This pub-sub is implemented in the `internal/broadcast` package. The publisher closes the gRPC connection. The subscribers call the cancel func associated with a context created for each streaming client.