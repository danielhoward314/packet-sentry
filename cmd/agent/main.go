package main

import (
	"github.com/danielhoward314/packet-sentry/agent"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
)

func initializeAgent(psAgent *agent.Agent) (err error) {
	// TODO: cert and install key logic here
	psAgent.BaseLogger.Info("instantiating agent dependencies")
	pcapManager := psPCap.NewPCapManager(psAgent.Ctx, psAgent.BaseLogger)
	// TODO: use goroutine and synchronization primitives instead of blocking
	psAgent.BaseLogger.Info("ensuring pcap manager is ready")
	err = pcapManager.EnsureReady()
	if err != nil {
		psAgent.BaseLogger.Error("failed pcap readiness check", psLog.KeyError, err)
		return err
	}
	psAgent.InjectDependencies(
		pcapManager,
	)
	return err
}
