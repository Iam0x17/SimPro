package services

import (
	"SimPro/config"
	"fmt"
	"strings"
	"sync"
)

type Service interface {
	Start(cfg *config.Config) error
	Stop() error
	GetName() string
}

type ServiceManager struct {
	services      map[string]Service
	serviceStatus map[string]string
	lock          sync.Mutex
	cfg           *config.Config
}

var manager *ServiceManager
var once sync.Once

func NewServiceManager(cfg *config.Config) *ServiceManager {
	once.Do(func() {
		manager = &ServiceManager{
			services:      make(map[string]Service),
			serviceStatus: make(map[string]string),
			cfg:           cfg,
		}
	})
	return manager
}

func GetServiceManager() *ServiceManager {
	return manager
}

func (s *ServiceManager) AddService(service Service) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.services[service.GetName()] = service
}

func (s *ServiceManager) StartServiceByName(serviceName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	lowServiceName := strings.ToLower(serviceName)
	service, exists := s.services[lowServiceName]
	if !exists {
		return fmt.Errorf("Service %s not found", serviceName)
	}
	err := service.Start(s.cfg)
	if err != nil {
		return err
	}
	s.serviceStatus[serviceName] = "Running"
	return nil
}

func (s *ServiceManager) StopServiceByName(serviceName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	service, exists := s.services[serviceName]
	if !exists {
		return fmt.Errorf("Service %s not found", serviceName)
	}
	err := service.Stop()
	if err != nil {
		return err
	}
	s.serviceStatus[serviceName] = "Stopped"
	return nil
}

func (s *ServiceManager) StartAllServices() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, service := range s.services {
		err := service.Start(s.cfg)
		if err != nil {
			return err
		}
		s.serviceStatus[service.GetName()] = "Running"
	}
	return nil
}

func (s *ServiceManager) GetServiceStatusByName(serviceName string) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	status, exists := s.serviceStatus[serviceName]
	if !exists {
		return "", fmt.Errorf("Service %s not found", serviceName)
	}
	return status, nil
}
