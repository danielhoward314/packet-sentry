//go:build darwin || linux

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/danielhoward314/packet-sentry/agent"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const (
	pidFileName = "/var/run/packetsentryagent.pid"
)

func main() {
	var err error

	psAgent := agent.NewAgent()
	if psAgent.BaseLogger == nil {
		panic("failed to get new agent instance")
	}

	psAgent.BaseLogger.Info("initializing agent")
	err = initializeAgent(psAgent)
	if err != nil {
		psAgent.BaseLogger.Error("failed to initialize agent", psLog.KeyError, err)
		panic(fmt.Sprintf("main: could not initialize agent due to %s", err))
	}

	err = os.WriteFile(pidFileName, []byte(strconv.Itoa(os.Getpid())), 0o644)
	if err != nil {
		psAgent.BaseLogger.Error("failed to write pid file", psLog.KeyError, err)
		panic(fmt.Sprintf("main: could not write pid file due to %s", err))
	}

	var wg sync.WaitGroup
	wg.Add(1)

	shutdownChan := make(chan struct{})

	// Start agent in background
	go func() {
		// only used for agent goroutine, as the signal one below has no cleanup associated with it
		defer wg.Done()
		agentStartErr := psAgent.Start()
		if agentStartErr != nil {
			psAgent.BaseLogger.Error("failed to start agent", psLog.KeyError, agentStartErr)
			psAgent.BaseLogger.Info("agent.Start() errored, calling agent.Stop()")
			psAgent.Stop()
			select {
			case shutdownChan <- struct{}{}:
			default:
			}
		}
		psAgent.BaseLogger.Info("main agent.Start is exiting")
	}()

	// Listen for OS signals in background
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
		receivedSignal := <-signalChannel
		psAgent.BaseLogger.Info("received sigterm or sigint, shutting down agent", slog.String("signal", receivedSignal.String()))
		psAgent.Stop()
		select {
		case shutdownChan <- struct{}{}:
		default:
		}
	}()

	// block until either signal or start failure
	<-shutdownChan
	// wait for the agent goroutine to be done
	wg.Wait()
	psAgent.BaseLogger.Info("agent shutdown complete")
}
