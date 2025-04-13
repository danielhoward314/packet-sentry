package poll

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/config"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
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
	cancelFunc            context.CancelFunc
	client                *http.Client
	clientMu              sync.RWMutex
	commandsBroadcaster   *broadcast.CommandsBroadcaster
	ctx                   context.Context
	logger                *slog.Logger
	mTLSClientBroadcaster *broadcast.MTLSClientBroadcaster
	pollInterval          time.Duration
	shutdownChannel       chan struct{}
	stopOnce              sync.Once
}

// NewPollManager returns an implementation of the PollManager interface
func NewPollManager(
	ctx context.Context,
	baseLogger *slog.Logger,
	commandsBroadcaster *broadcast.CommandsBroadcaster,
	mtlsClientBroadCaster *broadcast.MTLSClientBroadcaster,
) PollManager {
	childCtx, cancelFunc := context.WithCancel(ctx)
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))

	pm := &pollManager{
		cancelFunc:            cancelFunc,
		commandsBroadcaster:   commandsBroadcaster,
		ctx:                   childCtx,
		logger:                childLogger,
		mTLSClientBroadcaster: mtlsClientBroadCaster,
		pollInterval:          config.GetPollInterval(),
		shutdownChannel:       make(chan struct{}),
	}

	return pm
}

// Start runs an infinite loop in a goroutine, polling the server for commands on a configured interval
// and subscribing to mTLS client updates
func (pm *pollManager) Start() {
	logger := pm.logger.With(psLog.KeyFunction, "PollManager.Start")
	logger.Info("starting poll manager")

	sub := pm.mTLSClientBroadcaster.Subscribe()

	for {
		select {
		case clientUpdate := <-sub:
			pm.clientMu.Lock()
			pm.client = clientUpdate.Client
			pm.clientMu.Unlock()
		case <-time.After(pm.pollInterval):
			logger.Info("poll interval elapsed, sending GET to /poll")

			req, err := http.NewRequestWithContext(pm.ctx, http.MethodGet, "https://localhost:9443/poll", nil)
			if err != nil {
				logger.Error("failed to create request", "error", err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			logger.Info("sending http request", psLog.KeyURI, "/poll")
			pm.clientMu.RLock()
			client := pm.client
			pm.clientMu.RUnlock()
			if client == nil {
				logger.Error("no mTLS client available, skipping poll until next poll interval elapses")
				continue
			}

			resp, err := pm.client.Do(req)
			if err != nil {
				logger.Error("failed to poll server", "error", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Warn("received non-200 response from /poll", psLog.KeyStatus, resp.Status)
				continue
			}

			logger.Info("reading http response body")
			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("failed to read response body", psLog.KeyError, err)
				continue
			}

			logger.Info("unmarshaling http response body")
			var cmd broadcast.Command
			err = json.Unmarshal(resBody, &cmd)
			if err != nil {
				logger.Error("failed to unmarshal response body", psLog.KeyError, err)
				continue
			}

			if cmd.Name != "noop" {
				logger.Info("received command", psLog.KeyCommand, cmd.Name)
				pm.commandsBroadcaster.Publish(&cmd)
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
