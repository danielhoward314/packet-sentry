package main

import (
	"github.com/danielhoward314/packet-sentry/agent"
	ndrmgr "github.com/danielhoward314/packet-sentry/internal/pcap"
)

func initializeAgent(psAgent *agent.Agent) (err error) {
	// TODO: cert and install key logic here
	pcapManager := ndrmgr.NewPCapManager(psAgent.Ctx)
	// TODO: use goroutine and synchronization primitives instead of blocking
	err = pcapManager.EnsureReady()
	if err != nil {
		return err
	}
	psAgent.InjectDependencies(
		pcapManager,
	)
	return err
}
