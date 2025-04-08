package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielhoward314/packet-sentry/agent"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psOS "github.com/danielhoward314/packet-sentry/internal/os"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
	"github.com/danielhoward314/packet-sentry/internal/transport"
)

func initializeAgent(psAgent *agent.Agent) error {
	psAgent.BaseLogger.With(psLog.KeyFunction, "main.initializeAgent")

	psAgent.BaseLogger.Info("instantiating agent dependencies")
	systemInfo := psOS.NewSystemInfo(psAgent.Ctx, psAgent.BaseLogger)
	certManager := transport.NewCertificateManager(psAgent.Ctx, psAgent.BaseLogger, systemInfo)
	pcapManager := psPCap.NewPCapManager(psAgent.Ctx, psAgent.BaseLogger)

	certCh := make(chan error, 1)
	pcapCh := make(chan error, 1)
	timeout := 2 * time.Minute
	goRoutinesCount := 0

	goRoutinesCount++
	go func() {
		psAgent.BaseLogger.Info("ensuring client certificate is in place for mTLS")
		err := certManager.HasValidCert()
		if err != nil {
			isRenewal := false
			switch err.(type) {
			case *transport.CertExpiringSoonError:
				psAgent.BaseLogger.Warn("client certificate will expire within 30 days, requesting renewal")
				isRenewal = true
			default:
				psAgent.BaseLogger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
			}
			err = certManager.RequestCert(isRenewal)
		}
		certCh <- err
	}()

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

	var certErr, pcapErr error

	psAgent.BaseLogger.Info("processing goroutines with multi-select, will loop until finished or timeout", slog.Int("goroutinesCount", goRoutinesCount))
	done := 0
	for done < goRoutinesCount {
		select {
		case certErr = <-certCh:
			psAgent.BaseLogger.Info("certificate initialization completed", "error", certErr)
			done++
		case pcapErr = <-pcapCh:
			psAgent.BaseLogger.Info("pcap initialization completed", "error", pcapErr)
			done++
		case <-timer.C:
			return fmt.Errorf("timeout waiting for dependency initialization")
		}
	}

	if certErr != nil {
		return fmt.Errorf("certificate initialization failed: %w", certErr)
	}

	if pcapErr != nil {
		return fmt.Errorf("pcap initialization failed: %w", pcapErr)
	}

	psAgent.BaseLogger.Info("instantiating mTLS client")
	mTLSClient, err := certManager.GetMTLSClient()
	if err != nil {
		psAgent.BaseLogger.Error("failed to get mTLS client", psLog.KeyError, err)
		return err
	}

	psAgent.BaseLogger.Info("testing mTLS communication with server")
	// TODO: env-based config of base URL
	res, err := mTLSClient.Get("https://localhost:9443/mtls-test")
	if err != nil {
		psAgent.BaseLogger.Error("mTLS request failed", psLog.KeyError, err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("mTLS response status is non-200: %d", res.StatusCode)
		psAgent.BaseLogger.Error("failed mTLS readiness check", psLog.KeyError, err)
		return err
	}

	psAgent.BaseLogger.Info("mTLS client and server communication successful")

	psAgent.BaseLogger.Info("all dependencies ready, proceeding with injection")
	psAgent.InjectDependencies(
		certManager,
		mTLSClient,
		pcapManager,
	)

	return nil
}
