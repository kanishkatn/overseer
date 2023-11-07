package main

import (
	"context"
	logging "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"os/signal"
	"overseer/eventbus"
	"syscall"
	"testing"
)

// TestSubsystemLibrary tests the subsystem orchestration
func TestSubsystemLibrary(t *testing.T) {
	systemEventBus := eventbus.New()
	ctx, cancel := context.WithCancel(context.Background())
	systemContext := context.WithValue(ctx, "test", 1)

	subsystem1 := NewSubsystem1(systemContext, systemEventBus)
	subsystem2 := NewSubsystem2(systemContext, systemEventBus)

	overseer := NewOverseer(
		systemEventBus,
		subsystem1,
		subsystem2,
	)

	for name, baseSubsystem := range overseer.Subsystems {
		startSubsystem(t, name, baseSubsystem)
	}

	// serviceLibrary provides a way to call subsystem methods
	// this could be used by both the overseer and the subsystems
	subsystemLibrary := NewSubsystemLibrary(systemEventBus, "test")

	// ping subsystem1
	_, err := subsystemLibrary.Subsystem1Methods().Ping("hello")
	require.NoError(t, err)

	// ping subsystem1 from subsystem2
	_, err = subsystemLibrary.Subsystem2Methods().PingSubsystem1("hello")
	require.NoError(t, err)

	// process active leaves update in subsystem2
	_, err = subsystemLibrary.Subsystem2Methods().ProcessActiveLeavesUpdate()
	require.NoError(t, err)

	idleConnsClosed := make(chan struct{})

	// Stop upon receiving SIGTERM or CTRL-C
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		osSig := <-osSignal

		logging.Info("shutting down the node, received signal: " + osSig.String())
		cancel()

		// Exit the blocking chan
		close(idleConnsClosed)
	}()

	<-idleConnsClosed
}

func startSubsystem(t *testing.T, name string, subsystem *BaseSubsystem) {
	go func() {
		logging.WithField("name", name).Debug("baseSubsystem start")
		ok, err := subsystem.Start()
		require.NoError(t, err, "baseSubsystem start error")
		require.True(t, ok, "baseSubsystem start failed")
		logging.WithField("name", name).Debug("baseSubsystem started")
	}()
}
