package main

import (
	"fmt"
	"log"
	"net/http"

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

	// TODO: use goroutine and synchronization primitives instead of blocking
	psAgent.BaseLogger.Info("ensuring client certificate is in place for mTLS")
	err := certManager.HasValidCert()
	if err != nil {
		isRenewal := false
		var certReqErr error
		switch err.(type) {
		case *transport.CertExpiringSoonError:
			psAgent.BaseLogger.Warn("client certificate will expire within 30 days, requesting renewal")
			isRenewal = true
		default:
			psAgent.BaseLogger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
			isRenewal = false
		}
		certReqErr = certManager.RequestCert(isRenewal)
		if certReqErr != nil {
			return certReqErr
		}
	}

	psAgent.BaseLogger.Info("instantiating mTLS client")
	mTLSClient, err := certManager.GetMTLSClient()
	if err != nil {
		psAgent.BaseLogger.Error("failed to get mTLS client", psLog.KeyError, err)
	}

	psAgent.BaseLogger.Info("testing mTLS communication with server")
	// TODO: env-based config of base URL
	res, err := mTLSClient.Get("https://localhost:9443/mtls-test")
	if err != nil {
		log.Fatalf("mTLS request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("mTLS response status is non-200: %d", res.StatusCode)
		psAgent.BaseLogger.Error("failed mTLS readiness check", psLog.KeyError, err)
		return err
	}

	psAgent.BaseLogger.Info("ensuring pcap manager is ready")
	err = pcapManager.EnsureReady()
	if err != nil {
		psAgent.BaseLogger.Error("failed pcap readiness check", psLog.KeyError, err)
		return err
	}
	psAgent.InjectDependencies(
		certManager,
		mTLSClient,
		pcapManager,
	)
	return err
}
