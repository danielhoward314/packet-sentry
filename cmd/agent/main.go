package main

import (
	"time"

	"github.com/danielhoward314/packet-sentry/agent"
	"github.com/danielhoward314/packet-sentry/services/dummy"
)

func initializeAgent(psAgent *agent.Agent) (err error) {
	// TODO: cert and install key logic here
	psAgent.InjectDependencies(dummy.NewDummyService(1 * time.Minute))
	return err
}
