package agent

import (
	"context"
	"log/slog"
	"net/http"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
	"github.com/danielhoward314/packet-sentry/internal/transport"
)

type Agent struct {
	BaseLogger         *slog.Logger
	CertificateManager transport.CertificateManager
	Ctx                context.Context
	MTLSClient         *http.Client
	PCapManager        psPCap.PCapManager
}

func NewAgent() *Agent {
	return &Agent{
		BaseLogger: psLog.GetBaseLogger(),
		Ctx:        context.Background(),
	}
}

func (agent *Agent) InjectDependencies(
	certManager transport.CertificateManager,
	mTLSClient *http.Client,
	pcapManager psPCap.PCapManager,
) {
	agent.BaseLogger.With(psLog.KeyFunction, "Agent.InjectDependencies")
	agent.BaseLogger.Info("injecting agent dependencies")
	agent.CertificateManager = certManager
	agent.MTLSClient = mTLSClient
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
