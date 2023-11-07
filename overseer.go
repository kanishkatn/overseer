package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"overseer/eventbus"
	"sync"

	"github.com/avast/retry-go"
	logging "github.com/sirupsen/logrus"
)

const SECP256k1GeneratorOrder = "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"

type Overseer struct {
	eventBus      eventbus.Bus
	middlewareMap sync.Map
	Subsystems    map[string]*BaseSubsystem
}

func NewOverseer(eventBus eventbus.Bus, baseSubsystems ...*BaseSubsystem) *Overseer {
	overseer := &Overseer{
		Subsystems: make(map[string]*BaseSubsystem),
	}
	overseer.SetEventBus(eventBus)
	overseer.SetupMethodRouting()
	for _, baseSubsystem := range baseSubsystems {
		overseer.RegisterSubsystem(baseSubsystem)
	}
	return overseer
}

func (s *Overseer) AddMiddleware(middleware *func(string, interface{}) interface{}) {
	s.middlewareMap.Store(middleware, middleware)
	s.eventBus.AddMiddleware(middleware)
}

func (s *Overseer) RemoveMiddleware(m *func(string, interface{}) interface{}) {
	middleware, ok := s.middlewareMap.Load(m)
	if !ok {
		return
	}
	s.eventBus.RemoveMiddleware(middleware.(*func(string, interface{}) interface{}))
	s.middlewareMap.Delete(middleware)
}

func (s *Overseer) AddRequestMiddleware(middleware *func(methodRequest MethodRequest) MethodRequest) {
	requestMiddleware := func(topic string, inter interface{}) interface{} {
		if methodRequest, ok := inter.(MethodRequest); ok {
			return (*middleware)(methodRequest)
		}
		return inter
	}
	s.middlewareMap.Store(middleware, &requestMiddleware)
	s.eventBus.AddMiddleware(&requestMiddleware)
}

func (s *Overseer) AddResponseMiddleware(middleware *func(methodResponse MethodResponse) MethodResponse) {
	responseMiddleware := func(topic string, inter interface{}) interface{} {
		if methodResponse, ok := inter.(MethodResponse); ok {
			return (*middleware)(methodResponse)
		}
		return inter
	}
	s.middlewareMap.Store(middleware, &responseMiddleware)
	s.eventBus.AddMiddleware(&responseMiddleware)
}

func (s *Overseer) RemoveRequestMiddleware(middleware *func(methodRequest MethodRequest) MethodRequest) {
	requestMiddleware, ok := s.middlewareMap.Load(middleware)
	if !ok {
		return
	}
	s.eventBus.RemoveMiddleware(requestMiddleware.(*func(string, interface{}) interface{}))
	s.middlewareMap.Delete(middleware)
}

func (s *Overseer) RemoveResponseMiddleware(middleware *func(methodResponse MethodResponse) MethodResponse) {
	responseMiddleware, ok := s.middlewareMap.Load(middleware)
	if !ok {
		return
	}
	s.eventBus.RemoveMiddleware(responseMiddleware.(*func(string, interface{}) interface{}))
	s.middlewareMap.Delete(middleware)
}

func (s *Overseer) SetEventBus(e eventbus.Bus) {
	s.eventBus = e
}

func (s *Overseer) RegisterSubsystem(bs *BaseSubsystem) {
	s.Subsystems[bs.Name()] = bs
}

type MethodRequest struct {
	Caller    string
	Subsystem string
	Method    string
	ID        string
	Data      []interface{}
}

func (s *Overseer) SetupMethodRouting() {
	err := s.eventBus.SubscribeAsync("method", func(data interface{}) {
		methodRequest, ok := data.(MethodRequest)
		if !ok {
			logging.Error("could not parse data for query")
			return
		}

		var baseSubsystem *BaseSubsystem
		err := retry.Do(func() error {
			bs, ok := s.Subsystems[methodRequest.Subsystem]
			if !ok {
				logging.WithField("Subsystem", methodRequest.Subsystem).Error("could not find subsystem")
				return fmt.Errorf("could not find subsystem %v", methodRequest.Subsystem)
			}
			if !bs.IsRunning() {
				logging.WithFields(logging.Fields{
					"Subsystem": methodRequest.Subsystem,
					"data":      data,
				}).Error("Subsystem is not running")
				return fmt.Errorf("subsystem %v is not running", methodRequest.Subsystem)
			}
			baseSubsystem = bs
			return nil
		})
		if err != nil {
			s.eventBus.Publish(methodRequest.ID, MethodResponse{
				Error: err,
				Data:  nil,
			})
		} else {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logging.WithFields(logging.Fields{
							"Caller": methodRequest.Caller,
							"Method": methodRequest.Method,
							"data":   data,
							"error":  err,
						}).Error("panicked during baseSubsystem.Call")
						resp := MethodResponse{
							Error: fmt.Errorf("%v", err),
							Data:  nil,
						}
						s.eventBus.Publish(methodRequest.ID, resp)
					}
				}()
				data, err := baseSubsystem.Call(methodRequest.Method, methodRequest.Data...)
				resp := MethodResponse{
					Request: methodRequest,
					Error:   err,
					Data:    data,
				}
				s.eventBus.Publish(methodRequest.ID, resp)
			}()
		}
	}, false)
	if err != nil {
		logging.WithError(err).Error("could not subscribe async")
	}
}

type MethodResponse struct {
	Request MethodRequest
	Error   error
	Data    interface{}
}

func AwaitTopic(eventBus eventbus.Bus, topic string) <-chan interface{} {
	responseCh := make(chan interface{})
	err := eventBus.SubscribeOnceAsync(topic, func(res interface{}) {
		responseCh <- res
		close(responseCh)
	})
	if err != nil {
		logging.WithError(err).Error("could not subscribe async")
	}
	return responseCh
}

func SubsystemMethod(eventBus eventbus.Bus, caller string, subsystem string, method string, data ...interface{}) MethodResponse {
	generatorOrder := new(big.Int)
	_, ok := generatorOrder.SetString(SECP256k1GeneratorOrder, 0)
	if !ok {
		return MethodResponse{
			Error: errors.New("could not parse SECP256k1GeneratorOrder"),
			Data:  nil,
		}
	}

	nonce, err := rand.Int(rand.Reader, generatorOrder)
	if err != nil {
		return MethodResponse{
			Error: errors.New("could not generate random nonce"),
			Data:  nil,
		}
	}
	nonceStr := nonce.Text(16)
	responseCh := AwaitTopic(eventBus, nonceStr)
	eventBus.Publish("method", MethodRequest{
		Caller:    caller,
		Subsystem: subsystem,
		Method:    method,
		ID:        nonceStr,
		Data:      data,
	})
	methodResponseInter := <-responseCh
	methodResponse, ok := methodResponseInter.(MethodResponse)
	if !ok {
		return MethodResponse{
			Error: errors.New("method response was not of MethodResponse type"),
			Data:  nil,
		}
	}
	return methodResponse
}

func EmptyHandler(name string) func() {
	return func() {
		logging.WithField("name", name).Error("handler was not initialized")
	}
}
