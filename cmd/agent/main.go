package main

import (
	"github.com/danielhoward314/packet-sentry/agent"
	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/certs"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psOS "github.com/danielhoward314/packet-sentry/internal/os"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
	"github.com/danielhoward314/packet-sentry/internal/poll"
)

func initializeAgent(psAgent *agent.Agent) error {
	logger := psAgent.BaseLogger.With(psLog.KeyFunction, "main.initializeAgent")

	logger.Info("instantiating agent dependencies")
	systemInfo := psOS.NewSystemInfo(psAgent.Ctx, psAgent.BaseLogger)
	mTLSClientBroadcaster := broadcast.NewMTLSClientBroadcaster()
	commandsBroadcaster := broadcast.NewCommandsBroadcaster()
	certManager := certs.NewCertificateManager(psAgent.Ctx, psAgent.BaseLogger, systemInfo, mTLSClientBroadcaster)
	pcapManager := psPCap.NewPCapManager(psAgent.Ctx, psAgent.BaseLogger, commandsBroadcaster, mTLSClientBroadcaster)
	pollManager := poll.NewPollManager(psAgent.Ctx, psAgent.BaseLogger, commandsBroadcaster, mTLSClientBroadcaster)

	logger.Info("ensuring client certificate is in place for mTLS")
	err := certManager.Init()
	if err != nil {
		logger.Error("failed client certificate readiness check", psLog.KeyError, err)
		return err
	}

	logger.Info("all dependencies ready, proceeding with injection")
	psAgent.InjectDependencies(
		certManager,
		pcapManager,
		pollManager,
	)

	return nil
}
