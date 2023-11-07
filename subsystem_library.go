package main

import "overseer/eventbus"

// SubsystemLibrary a wrapper around the event bus to facilitate method calls to other subsystems.
type SubsystemLibrary interface {
	SetOwner(owner string)
	GetOwner() (owner string)
	GetEventBus() eventbus.Bus
	SetEventBus(eventbus.Bus)
	Subsystem1Methods() Subsystem1Methods
	Subsystem2Methods() Subsystem2Methods
}

type SubsystemLibraryInstance struct {
	eventBus eventbus.Bus
	owner    string
}

func NewSubsystemLibrary(eventBus eventbus.Bus, owner string) *SubsystemLibraryInstance {
	return &SubsystemLibraryInstance{eventBus: eventBus, owner: owner}
}

func (sL *SubsystemLibraryInstance) SetOwner(owner string) {
	sL.owner = owner
}

func (sL *SubsystemLibraryInstance) GetOwner() (owner string) {
	return sL.owner
}

func (sL *SubsystemLibraryInstance) GetEventBus() eventbus.Bus {
	return sL.eventBus
}

func (sL *SubsystemLibraryInstance) SetEventBus(eventBus eventbus.Bus) {
	sL.eventBus = eventBus
}

func (sL *SubsystemLibraryInstance) Subsystem1Methods() Subsystem1Methods {
	return &Subsystem1MethodsInstance{eventBus: sL.eventBus}
}

func (sL *SubsystemLibraryInstance) Subsystem2Methods() Subsystem2Methods {
	return &Subsystem2MethodsInstance{eventBus: sL.eventBus}
}

type Subsystem1Methods interface {
	SetOwner(owner string)
	GetOwner() (owner string)
	Ping(message string) (any, error)
}

type Subsystem1MethodsInstance struct {
	owner    string
	eventBus eventbus.Bus
}

func (s1 *Subsystem1MethodsInstance) GetOwner() (owner string) {
	return s1.owner
}

func (s1 *Subsystem1MethodsInstance) SetOwner(owner string) {
	s1.owner = owner
}

func (s1 *Subsystem1MethodsInstance) Ping(message string) (any, error) {
	methodResponse := SubsystemMethod(s1.eventBus, s1.owner, "subsystem1", "ping", message)
	if methodResponse.Error != nil {
		return nil, methodResponse.Error
	}

	return methodResponse.Data, nil
}

type Subsystem2Methods interface {
	SetOwner(owner string)
	GetOwner() (owner string)
	PingSubsystem1(message string) (any, error)
	ProcessActiveLeavesUpdate() (any, error)
}

type Subsystem2MethodsInstance struct {
	owner    string
	eventBus eventbus.Bus
}

func (s2 *Subsystem2MethodsInstance) GetOwner() (owner string) {
	return s2.owner
}

func (s2 *Subsystem2MethodsInstance) SetOwner(owner string) {
	s2.owner = owner
}

func (s2 *Subsystem2MethodsInstance) PingSubsystem1(message string) (any, error) {
	methodResponse := SubsystemMethod(s2.eventBus, s2.owner, "subsystem2", "ping_subsystem1", message)
	if methodResponse.Error != nil {
		return nil, methodResponse.Error
	}

	return methodResponse.Data, nil
}

func (s2 *Subsystem2MethodsInstance) ProcessActiveLeavesUpdate() (any, error) {
	methodResponse := SubsystemMethod(s2.eventBus, s2.owner, "subsystem2", "process_active_leaves_update")
	if methodResponse.Error != nil {
		return nil, methodResponse.Error
	}

	return methodResponse.Data, nil
}
