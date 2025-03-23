//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/danielhoward314/packet-sentry/agent"
)

const (
	pidFileName = "/var/run/packetsentryagent.pid"
)

func main() {
	var err error
	psAgent := agent.NewAgent()
	err = initializeAgent(psAgent)
	if err != nil {
		panic(fmt.Sprintf("main: could not initialize agent due to %s", err))
	}
	err = os.WriteFile(pidFileName, []byte(strconv.Itoa(os.Getpid())), 0o644)
	if err != nil {
		panic(fmt.Sprintf("main: could not write pid file due to %s", err))
	}
	go func() {
		err = psAgent.Start()
		if err != nil {
			panic(fmt.Sprintf("Agent failed to start due to %s", err))
		}
	}()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	// TODO: log the signal with a shutdown message
	_ = <-signalChannel
	psAgent.Stop()
}
