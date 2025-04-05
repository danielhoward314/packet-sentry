//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/danielhoward314/packet-sentry/agent"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const serviceName = "PacketSentryAgent"

var (
	elog      debug.Log
	isPaused  bool
	pauseLock sync.Mutex
)

type psService struct{}

// Execute handles Windows Service control requests
func (m *psService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// Notify Windows that the service is in the start-up phase
	s <- svc.Status{State: svc.StartPending}

	psAgent := agent.NewAgent()
	if psAgent.BaseLogger == nil {
		panic("failed to get new agent instance")
	}
	psAgent.BaseLogger.Info("initializing agent")

	initDone := make(chan error, 1)
	go func() {
		initDone <- initializeAgent(psAgent)
	}()

	select {
	case err := <-initDone:
		if err != nil {
			psAgent.BaseLogger.Error("initialization failed", psLog.KeyError, err)
			return false, 1
		}
	case <-time.After(25 * time.Second):
		psAgent.BaseLogger.Warn("initialization is slow, proceeding with startup")
	}

	// Mark service as running and accept stop/pause/continue
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
	psAgent.BaseLogger.Info("started Windows service", psLog.KeyServiceName, serviceName)

	// Start the agent in a goroutine
	go func() {
		err := psAgent.Start()
		if err != nil {
			psAgent.BaseLogger.Error("failed to start agent", psLog.KeyError, err)
			s <- svc.Status{State: svc.Stopped}
		}
	}()

	// Service Control Loop
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			psAgent.BaseLogger.Info("Windows Service Control Manager is interrogating service state")
			s <- c.CurrentStatus

		case svc.Stop, svc.Shutdown:
			psAgent.BaseLogger.Info("Windows Service Control Manager sent stop or shutdown, stopping service")
			psAgent.Stop()
			s <- svc.Status{State: svc.Stopped}
			return false, 0

		case svc.Pause:
			pauseLock.Lock()
			if !isPaused {
				isPaused = true
				s <- svc.Status{State: svc.Paused, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
				psAgent.BaseLogger.Info("Windows Service Control Manager is pausing the service")
			}
			pauseLock.Unlock()

		case svc.Continue:
			pauseLock.Lock()
			if isPaused {
				isPaused = false
				s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
				psAgent.BaseLogger.Info("Windows Service Control Manager is resuming the service")
			}
			pauseLock.Unlock()

		default:
			psAgent.BaseLogger.Warn("unexpected control request", "serviceControlRequest", c.Cmd)
		}
	}

	return false, 0
}

// main starts the service or runs in debug mode
func main() {
	if len(os.Args) < 2 {
		isService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("failed to determine if running as a service: %v", err)
		}

		if isService {
			// Open Windows Event Log for logging
			elog, err = eventlog.Open(serviceName)
			if err != nil {
				log.Fatalf("Failed to open event log: %v", err)
			}
			defer elog.Close()

			elog.Info(1, fmt.Sprintf("%s: Starting service...", serviceName))
			err = svc.Run(serviceName, &psService{})
			if err != nil {
				elog.Error(1, fmt.Sprintf("%s: Service failed: %v", serviceName, err))
			}
			os.Exit(2)
		}
	}

	var cmdErr error
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		cmdErr = debug.Run(serviceName, &psService{})
		if cmdErr != nil {
			elog.Error(1, fmt.Sprintf("%s: Service failed: %v", serviceName, cmdErr))
		}
		os.Exit(2)
	}
}
