package poll

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/config"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
)

const (
	logAttrValSvcName = "pollManager"
)

// PollManager manages polling the server
type PollManager interface {
	Start()
	Stop()
}

type pollManager struct {
	agentMTLSClient            pbAgent.AgentServiceClient
	agentMTLSClientBroadcaster *broadcast.AgentMTLSClientBroadcaster
	agentMTLSClientMu          sync.RWMutex
	cancelFunc                 context.CancelFunc
	commandsBroadcaster        *broadcast.CommandsBroadcaster
	ctx                        context.Context
	logger                     *slog.Logger
	pollInterval               time.Duration
	shutdownChannel            chan struct{}
	stopOnce                   sync.Once
}

// NewPollManager returns an implementation of the PollManager interface
func NewPollManager(
	ctx context.Context,
	baseLogger *slog.Logger,
	commandsBroadcaster *broadcast.CommandsBroadcaster,
	agentMTLSClientBroadcaster *broadcast.AgentMTLSClientBroadcaster,
) PollManager {
	childCtx, cancelFunc := context.WithCancel(ctx)
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))

	pm := &pollManager{
		agentMTLSClientBroadcaster: agentMTLSClientBroadcaster,
		cancelFunc:                 cancelFunc,
		commandsBroadcaster:        commandsBroadcaster,
		ctx:                        childCtx,
		logger:                     childLogger,
		pollInterval:               config.GetPollInterval(),
		shutdownChannel:            make(chan struct{}),
	}

	return pm
}

// Start runs an infinite loop in a goroutine, polling the server for commands on a configured interval
// and subscribing to mTLS client updates
func (pm *pollManager) Start() {
	logger := pm.logger.With(psLog.KeyFunction, "PollManager.Start")
	logger.Info("starting poll manager")

	sub := pm.agentMTLSClientBroadcaster.Subscribe()

	for {
		select {
		case clientUpdate := <-sub:
			pm.agentMTLSClientMu.Lock()
			pm.agentMTLSClient = pbAgent.NewAgentServiceClient(clientUpdate.ClientConn)
			pm.agentMTLSClientMu.Unlock()
		case <-time.After(pm.pollInterval):
			logger.Info("sending poll request")
			pm.agentMTLSClientMu.RLock()
			client := pm.agentMTLSClient
			pm.agentMTLSClientMu.RUnlock()

			if client == nil {
				logger.Error("no mTLS client available, skipping poll until next poll interval elapses")
				continue
			}

			cmd, err := client.PollCommand(pm.ctx, &pbAgent.Empty{})
			if err != nil {
				logger.Error("failed to get command on poll", psLog.KeyError, err)
			}

			if cmd.Name != "noop" {
				logger.Info("received command", psLog.KeyCommand, cmd.Name)
				pm.commandsBroadcaster.Publish(&broadcast.Command{Name: cmd.Name})
			} else {
				logger.Info("received noop command, skipping publish")
			}

		case <-pm.ctx.Done():
			logger.Error("poll manager context canceled")
			return
		}
	}
}

func (pm *pollManager) Stop() {
	logger := pm.logger.With(psLog.KeyFunction, "PollManager.Stop")

	pm.stopOnce.Do(func() {
		logger.Info("stopping poll manager")
		pm.cancelFunc()
	})
}
