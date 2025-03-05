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
	s.serviceStatus[service.GetName()] = "Stopped"
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

func (s *ServiceManager) GetServiceConfig(serviceName string) (map[string]interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	lowServiceName := strings.ToLower(serviceName)
	_, exists := s.services[lowServiceName]
	if !exists {
		return nil, fmt.Errorf("Service %s not found", serviceName)
	}

	switch lowServiceName {
	case "ssh":
		return map[string]interface{}{
			"port": s.cfg.SSH.Port,
			"user": s.cfg.SSH.User,
			"pass": s.cfg.SSH.Pass,
		}, nil
	case "redis":
		return map[string]interface{}{
			"port": s.cfg.Redis.Port,
			"pass": s.cfg.Redis.Pass,
		}, nil
	case "postgres":
		return map[string]interface{}{
			"port": s.cfg.Postgres.Port,
			"user": s.cfg.Postgres.User,
			"pass": s.cfg.Postgres.Pass,
		}, nil
	case "mysql":
		return map[string]interface{}{
			"port": s.cfg.MySql.Port,
			"user": s.cfg.MySql.User,
			"pass": s.cfg.MySql.Pass,
		}, nil
	case "telnet":
		return map[string]interface{}{
			"port": s.cfg.Telnet.Port,
			"user": s.cfg.Telnet.User,
			"pass": s.cfg.Telnet.Pass,
		}, nil
	case "ftp":
		return map[string]interface{}{
			"port": s.cfg.FTP.Port,
			"user": s.cfg.FTP.User,
			"pass": s.cfg.FTP.Pass,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown service %s", serviceName)
	}
}

func (s *ServiceManager) UpdateServiceConfig(serviceName string, newConfig map[string]interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	lowServiceName := strings.ToLower(serviceName)
	_, exists := s.services[lowServiceName]
	if !exists {
		return fmt.Errorf("Service %s not found", serviceName)
	}

	switch lowServiceName {
	case "ssh":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.SSH.Port = port
		}
		if user, ok := newConfig["user"].(string); ok {
			s.cfg.SSH.User = user
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.SSH.Pass = pass
		}
	case "redis":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.Redis.Port = port
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.Redis.Pass = pass
		}
	case "postgres":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.Postgres.Port = port
		}
		if user, ok := newConfig["user"].(string); ok {
			s.cfg.Postgres.User = user
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.Postgres.Pass = pass
		}
	case "mysql":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.MySql.Port = port
		}
		if user, ok := newConfig["user"].(string); ok {
			s.cfg.MySql.User = user
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.MySql.Pass = pass
		}
	case "telnet":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.Telnet.Port = port
		}
		if user, ok := newConfig["user"].(string); ok {
			s.cfg.Telnet.User = user
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.Telnet.Pass = pass
		}
	case "ftp":
		if port, ok := newConfig["port"].(string); ok {
			s.cfg.FTP.Port = port
		}
		if user, ok := newConfig["user"].(string); ok {
			s.cfg.FTP.User = user
		}
		if pass, ok := newConfig["pass"].(string); ok {
			s.cfg.FTP.Pass = pass
		}
	default:
		return fmt.Errorf("Unknown service %s", serviceName)
	}

	return nil
}
