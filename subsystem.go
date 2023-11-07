package main

import (
	"sync/atomic"

	logging "github.com/sirupsen/logrus"
)

type Event string

const (
	// ErrorEvent is the event that is published when an error occurs.
	ErrorEvent Event = "error"

	// StartEvent is the event that is published when a subsystem is started.
	StartEvent Event = "start"

	// StopEvent is the event that is published when a subsystem is stopped.
	StopEvent Event = "stop"
)

// Subsystem is an interface that is implemented by a subsystem.
type Subsystem interface {
	Name() string
	OnStart() error
	OnStop() error
	Call(method string, args ...any) (result any, err error)
	SetBaseSubsystem(*BaseSubsystem)
}

// BaseSubsystem provides the guarantees that a Subsystem can only be started and stopped once.
type BaseSubsystem struct {
	name    string
	start   uint32 // atomic
	started uint32 // atomic
	stopped uint32 // atomic
	quit    chan struct{}

	// The "subclass" of BaseSubsystem
	impl Subsystem
}

// NewBaseSubsystem returns a BaseSubsystem that wraps an implementation of Subsystem and handles
// starting and stopping.
func NewBaseSubsystem(impl Subsystem) *BaseSubsystem {
	bs := &BaseSubsystem{
		name: impl.Name(),
		quit: make(chan struct{}),
		impl: impl,
	}
	bs.impl.SetBaseSubsystem(bs)
	return bs
}

// Name returns the name of the subsystem.
func (bs *BaseSubsystem) Name() string {
	return bs.impl.Name()
}

// Call calls a method on the subsystem.
func (bs *BaseSubsystem) Call(method string, args ...any) (any, error) {
	return bs.impl.Call(method, args...)
}

// Start starts the subsystem.
func (bs *BaseSubsystem) Start() (bool, error) {
	if atomic.CompareAndSwapUint32(&bs.start, 0, 1) {
		if atomic.LoadUint32(&bs.stopped) == 1 {
			logging.WithField("bsname", bs.name).Info("not starting basesubsystem -- already stopped")
			return false, nil
		} else {
			logging.WithField("bsname", bs.name).Info("starting subsystem")
		}
		err := bs.impl.OnStart()
		if err != nil {
			// revert flag
			atomic.StoreUint32(&bs.start, 0)
			return false, err
		}
		atomic.StoreUint32(&bs.started, 1)
		return true, err
	} else {
		logging.WithField("bsname", bs.name).Debug("not starting basesubsystem -- already stopped")
		return false, nil
	}
}

// Stop stops the subsystem.
func (bs *BaseSubsystem) Stop() bool {
	if atomic.CompareAndSwapUint32(&bs.stopped, 0, 1) {
		logging.WithField("bsname", bs.name).Info("stopping subsystem")
		err := bs.impl.OnStop()
		if err != nil {
			logging.WithField("bsname", bs.impl.Name()).WithError(err).Error("could not stop basesubsystem")
		}
		close(bs.quit)
		return true
	} else {
		logging.WithField("bsname", bs.name).Debug("stopping subsystem (ignoring: already stopped)")
		return false
	}
}

// IsRunning returns true if the subsystem is running.
func (bs *BaseSubsystem) IsRunning() bool {
	return atomic.LoadUint32(&bs.started) == 1 && atomic.LoadUint32(&bs.stopped) == 0
}

// SetSubsystem sets the implementation of the subsystem.
func (bs *BaseSubsystem) SetSubsystem(subsystem Subsystem) {
	bs.impl = subsystem
}

// String returns the name of the subsystem.
func (bs *BaseSubsystem) String() string {
	return bs.name
}

// Wait blocks until the subsystem is stopped.
func (bs *BaseSubsystem) Wait() {
	<-bs.quit
}
