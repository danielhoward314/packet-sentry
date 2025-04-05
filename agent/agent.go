package agent

import (
	"context"
	"log/slog"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
)

type Agent struct {
	BaseLogger  *slog.Logger
	Ctx         context.Context
	PCapManager psPCap.PCapManager
}

func NewAgent() *Agent {
	return &Agent{
		BaseLogger: psLog.GetBaseLogger(),
		Ctx:        context.Background(),
	}
}

func (agent *Agent) InjectDependencies(pcapManager psPCap.PCapManager) {
	agent.BaseLogger.With(psLog.KeyFunction, "Agent.InjectDependencies")
	agent.BaseLogger.Info("injecting agent dependencies")
	agent.PCapManager = pcapManager
}

func (agent *Agent) Start() (err error) {
	agent.BaseLogger.With(psLog.KeyFunction, "Agent.Start")
	agent.BaseLogger.Info("starting service managers")
	agent.PCapManager.StartAll()
	return
}

func (agent *Agent) Stop() (err error) {
	agent.BaseLogger.With(psLog.KeyFunction, "Agent.Stop")
	agent.BaseLogger.Info("stopping service managers")
	agent.PCapManager.StopAll()
	return
}
