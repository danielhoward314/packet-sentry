package agent

import (
	"github.com/danielhoward314/packet-sentry/services/dummy"
)

type Agent struct {
	DummyService dummy.DummyService
}

func NewAgent() *Agent {
	return &Agent{}
}

func (agent *Agent) InjectDependencies(dummyService dummy.DummyService) {
	agent.DummyService = dummyService
}

func (agent *Agent) Start() (err error) {
	agent.DummyService.Start()
	return
}

func (agent *Agent) Stop() (err error) {
	agent.DummyService.Stop()
	return
}
