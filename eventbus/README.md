# EventBus Package for Go

## Introduction

EventBus is a lightweight event bus in Go for asynchronous event handling with support for multiple event subscribers.

## Features

- Synchronous and asynchronous message publication
- Support for one-time or persistent subscriptions
- Middleware support for message interception
- Transactional event processing

## Quick Start

### Creating a new EventBus

```go
eb := eventbus.New()
```

### Subscribing to Events

```go
eb.Subscribe("topic:example", func(data any) {
    fmt.Printf("Received: %v\n", data)
})
```

### Publishing Events

```go
eb.Publish("topic:example", "Hello, World!")
```

### Unsubscribing from Events

```go
eb.Unsubscribe("topic:example", handler)
```

### Waiting for Asynchronous Events

```go
eb.WaitAsync()
```