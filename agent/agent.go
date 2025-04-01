package agent

import (
	"context"

	ndrmgr "github.com/danielhoward314/packet-sentry/internal/pcap"
)

type Agent struct {
	Ctx         context.Context
	PCapManager ndrmgr.PCapManager
}

func NewAgent() *Agent {
	return &Agent{
		Ctx: context.Background(),
	}
}

func (agent *Agent) InjectDependencies(pcapManager ndrmgr.PCapManager) {
	agent.PCapManager = pcapManager
}

func (agent *Agent) Start() (err error) {
	agent.PCapManager.StartAll()
	return
}

func (agent *Agent) Stop() (err error) {
	agent.PCapManager.StopAll()
	return
}
