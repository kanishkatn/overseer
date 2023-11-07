package main

import (
	"context"
	"errors"
	"fmt"
	logging "github.com/sirupsen/logrus"
	"overseer/eventbus"
)

type Subsystem1 struct {
	bs               *BaseSubsystem
	cancel           context.CancelFunc
	ctx              context.Context
	eventBus         eventbus.Bus
	subsystemLibrary SubsystemLibrary
}

func (t *Subsystem1) Name() string {
	return "subsystem1"
}

func (t *Subsystem1) OnStart() error {
	go func() {
		select {}
	}()
	t.eventBus.Publish(fmt.Sprintf("subsystem1:%v", StartEvent), nil)
	logging.Info("subsystem1 started")
	return nil
}

func (t *Subsystem1) OnStop() error {
	t.cancel()
	t.eventBus.Publish(fmt.Sprintf("subsystem1:%v", ErrorEvent), nil)
	return nil
}

func (t *Subsystem1) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "ping":
		logging.WithField("args", args).Info("subsystem1 ping called")
		return "pong", nil

	default:
		return nil, errors.New("method not found")
	}
}

func (t *Subsystem1) SetBaseSubsystem(bs *BaseSubsystem) {
	t.bs = bs
}

func NewSubsystem1(ctx context.Context, eventBus eventbus.Bus) *BaseSubsystem {
	ctx, cancel := context.WithCancel(context.WithValue(ctx, "subsystem", "subsystem1"))
	subsystem1 := Subsystem1{
		cancel:   cancel,
		ctx:      ctx,
		eventBus: eventBus,
	}
	subsystem1.subsystemLibrary = NewSubsystemLibrary(subsystem1.eventBus, subsystem1.Name())
	return NewBaseSubsystem(&subsystem1)
}
