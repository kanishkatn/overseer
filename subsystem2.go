package main

import (
	"context"
	"errors"
	"fmt"
	logging "github.com/sirupsen/logrus"
	"overseer/eventbus"
)

type Subsystem2 struct {
	bs               *BaseSubsystem
	cancel           context.CancelFunc
	ctx              context.Context
	eventBus         eventbus.Bus
	subsystemLibrary SubsystemLibrary
}

func (t *Subsystem2) Name() string {
	return "subsystem2"
}

func (t *Subsystem2) OnStart() error {
	go func() {
		select {}
	}()
	t.eventBus.Publish(fmt.Sprintf("subsystem2:%v", StartEvent), nil)
	return nil
}

func (t *Subsystem2) OnStop() error {
	t.cancel()
	t.eventBus.Publish(fmt.Sprintf("subsystem2:%v", ErrorEvent), nil)
	return nil
}

func (t *Subsystem2) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "ping":
		logging.WithField("args", args).Info("subsystem2 ping called")
		return "pong", nil

	case "ping_subsystem1":
		logging.WithField("args", args).Info("subsystem2 ping_subsystem1 called")
		if len(args) != 1 {
			return nil, errors.New("invalid number of args")
		}
		arg, ok := args[0].(string)
		if !ok {
			return nil, errors.New("invalid arg type")
		}

		return t.subsystemLibrary.Subsystem1Methods().Ping(arg)

	case "process_active_leaves_update":
		logging.WithField("args", args).Info("subsystem2 process_active_leaves_update called")
		return nil, nil

	default:
		return nil, errors.New("method not found")
	}
}

func (t *Subsystem2) SetBaseSubsystem(bs *BaseSubsystem) {
	t.bs = bs
}

func NewSubsystem2(ctx context.Context, eventBus eventbus.Bus) *BaseSubsystem {
	ctx, cancel := context.WithCancel(context.WithValue(ctx, "subsystem", "subsystem2"))
	subsystem2 := Subsystem2{
		cancel:   cancel,
		ctx:      ctx,
		eventBus: eventBus,
	}
	subsystem2.subsystemLibrary = NewSubsystemLibrary(subsystem2.eventBus, subsystem2.Name())
	return NewBaseSubsystem(&subsystem2)
}
