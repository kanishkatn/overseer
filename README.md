# Overseer Demo Project

## Introduction

The Overseer Demo Project showcases an event-driven system that orchestrates the interaction between different subsystems using an event bus. `Subsystem1` and `Subsystem2` are demo implementations that exhibit how subsystems can communicate, start, and stop within a broader application context.

## System Components

### Event Bus (`eventbus.Bus`)
The event bus serves as the communication backbone of the system, allowing for publish-subscribe interactions between subsystems.

**Example**: When `Subsystem1` wants to notify the system that it has started, it can publish a `StartEvent`:

```go
t.eventBus.Publish(fmt.Sprintf("subsystem1:%v", StartEvent), nil)
```

Any subsystem interested in `StartEvent` can subscribe to it and handle it accordingly. A good usecase here is `processActiveLeafUpdates` event that needs to be sent to all the subsystems.

### Subsystem Implementations (`Subsystem1` & `Subsystem2`)
These are concrete implementations of subsystems that perform specific tasks and communicate with other parts of the system using the event bus.

**Example**: `Subsystem1` responds to a "ping" request with "pong":

```go
func (t *Subsystem1) Call(method string, args ...interface{}) (interface{}, error) {
    switch method {
    case "ping":
        return "pong", nil
    // other cases...
    }
}
```

### Subsystem Library (`SubsystemLibrary`)
A higher-level abstraction that provides a user-friendly interface for invoking methods on subsystems without directly dealing with event bus messaging.

**Example**: To ping `Subsystem1` using the `SubsystemLibrary`:

```go
subsystemLibrary.Subsystem1Methods().Ping("Hello, Subsystem1!")
```

Behind the scenes, this uses the event bus to send the request to `Subsystem1`.

### Base Subsystem (`BaseSubsystem`)
Ensures that each subsystem follows a consistent lifecycle, only allowing it to be started and stopped once.

**Example**: Starting `Subsystem1` will set the appropriate flags to prevent it from being started again:

```go
ok, err := subsystem1.Start()
```

The `Start()` method uses atomic operations to manage the state safely in a concurrent environment.

### Context and CancelFunc
Used to manage the lifecycle of go-routines within subsystems, providing a way to gracefully stop operations and clean up resources.

**Example**: When `Subsystem2` is stopped, it signals through the context to terminate any ongoing operations:

```go
func (t *Subsystem2) OnStop() error {
    t.cancel()
    return nil
}
```

By calling `t.cancel()`, any context-aware operations within `Subsystem2` will be notified to stop.

### Integration of Components
All these components work together to create a flexible and maintainable system. For instance, when the overseer starts, it initializes the subsystems and uses the event bus to coordinate their actions:

```go
// Create subsystems
subsystem1 := NewSubsystem1(ctx, eventBus)
subsystem2 := NewSubsystem2(ctx, eventBus)

// Register subsystems with overseer
overseer := NewOverseer(eventBus, subsystem1, subsystem2)

// Start all registered subsystems
for _, baseSubsystem := range overseer.Subsystems {
    baseSubsystem.Start()
}
```

In this system, the overseer acts as the orchestrator, initializing subsystems and facilitating their communication through the event bus, 
all while ensuring that the subsystems are independently managed and can call one another, abstracting away the details of the event bus.

## Quick Start


### Running the Demo

To start the overseer and subsystems, execute the following steps:

```go
package main

import (
	"context"
	"overseer/eventbus"
)

func main() {
	// Initialize the event bus and context.
	systemEventBus := eventbus.New()
	ctx := context.Background()

	// Create subsystems with the event bus and context.
	subsystem1 := NewSubsystem1(ctx, systemEventBus)
	subsystem2 := NewSubsystem2(ctx, systemEventBus)

	// Initialize the overseer with the subsystems.
	overseer := NewOverseer(systemEventBus, subsystem1, subsystem2)

	// Start the subsystems.
	for _, baseSubsystem := range overseer.Subsystems {
		baseSubsystem.Start()
	}

	// Ping the subsystems.
	subsystemLibrary := NewSubsystemLibrary(systemEventBus, "DemoOverseer")
	subsystemLibrary.Subsystem1Methods().Ping("Hello!")

	// Stop the subsystems.
	for _, baseSubsystem := range overseer.Subsystems {
		baseSubsystem.Stop()
	}
}
```

### Testing

To run the included tests and validate the functionality:

```sh
go test -v ./...
```