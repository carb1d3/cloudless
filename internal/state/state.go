package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Service represents a managed service
type Service struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Status      string            `json:"status"` // created, running, stopped, failed
	ContainerID string            `json:"container_id,omitempty"`
	Port        int               `json:"port,omitempty"`
	Config      map[string]string `json:"config"`
	AutoStart   bool              `json:"auto_start"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// State manages service state persistence
type State struct {
	dir      string
	services map[string]*Service
	mu       sync.RWMutex
}

// New creates a new state manager
func New(dir string) (*State, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	s := &State{
		dir:      dir,
		services: make(map[string]*Service),
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

// SaveService saves a service to state
func (s *State) SaveService(svc *Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.services[svc.Name] = svc
	return s.persist()
}

// GetService retrieves a service from state
func (s *State) GetService(name string) (*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, ok := s.services[name]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	// Return a copy to prevent external modifications
	svcCopy := *svc
	return &svcCopy, nil
}

// ListServices returns all services
func (s *State) ListServices() ([]*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		svcCopy := *svc
		services = append(services, &svcCopy)
	}

	return services, nil
}

// DeleteService removes a service from state
func (s *State) DeleteService(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.services[name]; !ok {
		return fmt.Errorf("service '%s' not found", name)
	}

	delete(s.services, name)
	return s.persist()
}

// UpdateServiceStatus updates the status of a service
func (s *State) UpdateServiceStatus(name, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	svc, ok := s.services[name]
	if !ok {
		return fmt.Errorf("service '%s' not found", name)
	}

	svc.Status = status
	return s.persist()
}

// load loads state from disk
func (s *State) load() error {
	stateFile := filepath.Join(s.dir, "state.json")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var services []*Service
	if err := json.Unmarshal(data, &services); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	for _, svc := range services {
		s.services[svc.Name] = svc
	}

	return nil
}

// persist saves state to disk
func (s *State) persist() error {
	services := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, svc)
	}

	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	stateFile := filepath.Join(s.dir, "state.json")
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
