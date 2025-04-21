package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/danielhoward314/packet-sentry/internal/certs"
	"github.com/danielhoward314/packet-sentry/internal/config"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psPCap "github.com/danielhoward314/packet-sentry/internal/pcap"
	"github.com/danielhoward314/packet-sentry/internal/poll"
)

type Agent struct {
	AgentAddr          string
	BaseLogger         *slog.Logger
	BootstrapAddr      string
	CancelFunc         context.CancelFunc
	CertificateManager certs.CertificateManager
	Ctx                context.Context
	PollManager        poll.PollManager
	PCapManager        psPCap.PCapManager
	stopOnce           sync.Once
}

type configJSON struct {
	AgentHost     string `json:"agentHost"`
	AgentPort     string `json:"agentPort"`
	BootstrapPort string `json:"bootstrapPort"`
}

// NewAgent creates a new agent instance with the minimum required dependencies.
func NewAgent() *Agent {
	ctx, cancelFunc := context.WithCancel(context.Background())
	baseLogger := psLog.GetBaseLogger()

	// default to prod values, read from optional config.json file for local dev
	agentAddr := "agent-api.packet-sentry:9444"
	bootstrapAddr := "agent-api.packet-sentry:9443"

	configFilePath := config.GetConfigFilePath()
	content, err := os.ReadFile(configFilePath)
	if err == nil {
		baseLogger.Info("reading config file")
		var cfg configJSON
		err = json.Unmarshal(content, &cfg)
		if err == nil {
			agentHost := strings.TrimSpace(cfg.AgentHost)
			agentPort := strings.TrimSpace(cfg.AgentPort)
			bootstrapPort := strings.TrimSpace(cfg.BootstrapPort)
			if agentHost != "" && agentPort != "" && bootstrapPort != "" {
				baseLogger.Info("setting agent and bootstrap addresses from config file")
				agentAddr = fmt.Sprintf("%s:%s", agentHost, agentPort)
				bootstrapAddr = fmt.Sprintf("%s:%s", agentHost, bootstrapPort)
			}
		}
	}

	return &Agent{
		AgentAddr:     agentAddr,
		BaseLogger:    baseLogger,
		BootstrapAddr: bootstrapAddr,
		CancelFunc:    cancelFunc,
		Ctx:           ctx,
	}
}

// InjectDependencies injects dependencies into the agent, including all managers.
func (agent *Agent) InjectDependencies(
	certManager certs.CertificateManager,
	pcapManager psPCap.PCapManager,
	pollManager poll.PollManager,
) {
	logger := agent.BaseLogger.With(psLog.KeyFunction, "Agent.InjectDependencies")
	logger.Info("injecting agent dependencies")
	agent.CertificateManager = certManager
	agent.PollManager = pollManager
	agent.PCapManager = pcapManager
}

// Start is called to start the goroutines of all of the managers.
func (agent *Agent) Start() (err error) {
	logger := agent.BaseLogger.With(psLog.KeyFunction, "Agent.Start")

	logger.Info("starting service managers")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.CertificateManager.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.PCapManager.StartAll()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.PollManager.Start()
	}()

	// block until agent's context is canceled
	<-agent.Ctx.Done()
	logger.Info("agent context canceled, shutting down managers")
	// Wait for all manager goroutines to exit
	wg.Wait()

	return nil
}

// Stop is called to stop the agent either when we error during startup or when the OS tells us to shut down (Unix signal / Windows Service Control Manager).
// It uses sync.Once to ensure just-once calls to the top-level agent context's cancel func and to each of the managers stop methods.
func (agent *Agent) Stop() {
	logger := agent.BaseLogger.With(psLog.KeyFunction, "Agent.Stop")

	agent.stopOnce.Do(func() {
		logger.Info("shutting down")
		agent.CancelFunc()
		agent.CertificateManager.Stop()
		agent.PCapManager.StopAll()
		agent.PollManager.Stop()
	})
}
