//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

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

	err := initializeAgent(psAgent)
	if err != nil {
		psAgent.BaseLogger.Error("failed to initialize agent", psLog.KeyError, err)
		panic(fmt.Sprintf("main: could not initialize agent due to %s", err))
	}

	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
	psAgent.BaseLogger.Info("started Windows service", psLog.KeyServiceName, serviceName)

	var wg sync.WaitGroup
	wg.Add(1)

	shutdownChan := make(chan struct{})

	// Start agent in background
	go func() {
		defer wg.Done()
		err := psAgent.Start()
		if err != nil {
			psAgent.BaseLogger.Error("failed to start agent", psLog.KeyError, err)
			psAgent.BaseLogger.Info("agent.Start() errored, calling agent.Stop()")
			psAgent.Stop()
			select {
			case shutdownChan <- struct{}{}:
			default:
			}
		}
		psAgent.BaseLogger.Info("main agent.Start is exiting")
	}()

	// Service Control Loop
	go func() {
		for c := range r {
			switch c.Cmd {
			case svc.Interrogate:
				psAgent.BaseLogger.Info("Windows SCM is interrogating service state")
				s <- c.CurrentStatus

			case svc.Stop, svc.Shutdown:
				psAgent.BaseLogger.Info("Windows SCM sent stop/shutdown, stopping agent")
				psAgent.Stop()
				select {
				case shutdownChan <- struct{}{}:
				default:
				}
				return

			case svc.Pause:
				pauseLock.Lock()
				if !isPaused {
					isPaused = true
					s <- svc.Status{State: svc.Paused, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
					psAgent.BaseLogger.Info("pausing service")
				}
				pauseLock.Unlock()

			case svc.Continue:
				pauseLock.Lock()
				if isPaused {
					isPaused = false
					s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue}
					psAgent.BaseLogger.Info("resuming service")
				}
				pauseLock.Unlock()

			default:
				psAgent.BaseLogger.Warn("unexpected control request", "serviceControlRequest", c.Cmd)
			}
		}
	}()

	// block until either svc.Stop|svc.Shutdown (Windows SCM) or agent start failure
	<-shutdownChan

	// Wait for agent goroutine to complete
	wg.Wait()
	psAgent.BaseLogger.Info("agent shutdown complete")
	s <- svc.Status{State: svc.Stopped}
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
