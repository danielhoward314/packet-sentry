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
	elog.Info(1, fmt.Sprintf("%s: Service is starting...", serviceName))

	psAgent := agent.NewAgent()

	// Notify Windows that the service is in the start-up phase
	s <- svc.Status{State: svc.StartPending}

	// Run initialization (pre-requisite work)
	initDone := make(chan error, 1)
	go func() {
		initDone <- initializeAgent(psAgent)
	}()

	select {
	case err := <-initDone:
		if err != nil {
			elog.Error(1, fmt.Sprintf("%s: Initialization failed: %s", serviceName, err))
			return false, 1
		}
	case <-time.After(25 * time.Second):
		elog.Warning(1, fmt.Sprintf("%s: Initialization is slow, continuing startup...", serviceName))
	}

	// Mark service as running and accept stop/pause/continue
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
	elog.Info(1, fmt.Sprintf("%s: Service is now running.", serviceName))

	// Start the agent in a goroutine
	go func() {
		err := psAgent.Start()
		if err != nil {
			elog.Error(1, fmt.Sprintf("Agent failed to start: %s", err))
			s <- svc.Status{State: svc.Stopped}
		}
	}()

	// Service Control Loop
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			// Windows SCM is checking service status
			s <- c.CurrentStatus

		case svc.Stop, svc.Shutdown:
			elog.Info(1, fmt.Sprintf("%s: Service stopping...", serviceName))
			psAgent.Stop()
			s <- svc.Status{State: svc.Stopped}
			return false, 0

		case svc.Pause:
			pauseLock.Lock()
			if !isPaused {
				isPaused = true
				s <- svc.Status{State: svc.Paused, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
				elog.Info(1, fmt.Sprintf("%s: Service paused.", serviceName))
			}
			pauseLock.Unlock()

		case svc.Continue:
			pauseLock.Lock()
			if isPaused {
				isPaused = false
				s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
				elog.Info(1, fmt.Sprintf("%s: Service resumed.", serviceName))
			}
			pauseLock.Unlock()

		default:
			elog.Warning(1, fmt.Sprintf("%s: Unexpected control request: %d", serviceName, c.Cmd))
		}
	}

	return false, 0
}

// main starts the service or runs in debug mode
func main() {
	if len(os.Args) < 2 {
		isService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("Failed to determine if running as a service: %v", err)
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
