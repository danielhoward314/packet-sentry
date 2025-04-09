package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/danielhoward314/packet-sentry/agent"
	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/certs"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psOS "github.com/danielhoward314/packet-sentry/internal/os"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
)

func initializeAgent(psAgent *agent.Agent) error {
	psAgent.BaseLogger.With(psLog.KeyFunction, "main.initializeAgent")

	psAgent.BaseLogger.Info("instantiating agent dependencies")
	systemInfo := psOS.NewSystemInfo(psAgent.Ctx, psAgent.BaseLogger)
	mTLSClientBroadcaster := broadcast.NewMTLSClientBroadcaster()
	certManager := certs.NewCertificateManager(psAgent.Ctx, psAgent.BaseLogger, systemInfo, mTLSClientBroadcaster)
	pcapManager := psPCap.NewPCapManager(psAgent.Ctx, psAgent.BaseLogger, mTLSClientBroadcaster)

	pcapCh := make(chan error, 1)
	timeout := 2 * time.Minute
	goRoutinesCount := 0

	psAgent.BaseLogger.Info("ensuring client certificate is in place for mTLS")
	err := certManager.Init()
	if err != nil {
		psAgent.BaseLogger.Error("failed client certificate readiness check", psLog.KeyError, err)
	}

	goRoutinesCount++
	go func() {
		psAgent.BaseLogger.Info("ensuring pcap manager is ready")
		err := pcapManager.EnsureReady()
		if err != nil {
			psAgent.BaseLogger.Error("failed pcap readiness check", psLog.KeyError, err)
		}
		pcapCh <- err
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var pcapErr error

	psAgent.BaseLogger.Info("processing goroutines with multi-select, will loop until finished or timeout", slog.Int("goroutinesCount", goRoutinesCount))
	done := 0
	for done < goRoutinesCount {
		select {
		case pcapErr = <-pcapCh:
			psAgent.BaseLogger.Info("pcap initialization completed", "error", pcapErr)
			done++
		case <-timer.C:
			return fmt.Errorf("timeout waiting for dependency initialization")
		}
	}

	if pcapErr != nil {
		return fmt.Errorf("pcap initialization failed: %w", pcapErr)
	}

	psAgent.BaseLogger.Info("all dependencies ready, proceeding with injection")
	psAgent.InjectDependencies(
		certManager,
		pcapManager,
	)

	return nil
}
