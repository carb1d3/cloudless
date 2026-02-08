package service

import (
	"context"
	"fmt"
	"time"

	"github.com/carb1d3/cloudless/internal/container"
	"github.com/carb1d3/cloudless/internal/state"
)

// Manager manages services
type Manager struct {
	state        *state.State
	containerMgr *container.Manager
	providers    map[string]Provider
}

// Provider defines the interface for service providers
type Provider interface {
	Create(ctx context.Context, name string) (*state.Service, error)
	Start(ctx context.Context, svc *state.Service) error
	Stop(ctx context.Context, svc *state.Service) error
	HealthCheck(ctx context.Context, svc *state.Service) (bool, error)
	Delete(ctx context.Context, svc *state.Service) error
}

// NewManager creates a new service manager
func NewManager(st *state.State, containerMgr *container.Manager) *Manager {
	m := &Manager{
		state:        st,
		containerMgr: containerMgr,
		providers:    make(map[string]Provider),
	}

	// Register providers
	m.providers["postgres"] = NewPostgresProvider(containerMgr)

	return m
}

// Create creates a new service
func (m *Manager) Create(ctx context.Context, serviceType, name string) error {
	// Check if service already exists
	if _, err := m.state.GetService(name); err == nil {
		return fmt.Errorf("service '%s' already exists", name)
	}

	// Get provider
	provider, ok := m.providers[serviceType]
	if !ok {
		return fmt.Errorf("unsupported service type: %s", serviceType)
	}

	// Create service
	svc, err := provider.Create(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Save to state
	if err := m.state.SaveService(svc); err != nil {
		return fmt.Errorf("failed to save service state: %w", err)
	}

	// Start the service
	if err := m.Start(ctx, name); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

// Start starts a service
func (m *Manager) Start(ctx context.Context, name string) error {
	svc, err := m.state.GetService(name)
	if err != nil {
		return err
	}

	if svc.Status == "running" {
		return fmt.Errorf("service '%s' is already running", name)
	}

	provider, ok := m.providers[svc.Type]
	if !ok {
		return fmt.Errorf("unsupported service type: %s", svc.Type)
	}

	if err := provider.Start(ctx, svc); err != nil {
		svc.Status = "failed"
		m.state.SaveService(svc)
		return fmt.Errorf("failed to start service: %w", err)
	}

	svc.Status = "running"
	svc.UpdatedAt = time.Now().Format(time.RFC3339)
	return m.state.SaveService(svc)
}

// Stop stops a service
func (m *Manager) Stop(ctx context.Context, name string) error {
	svc, err := m.state.GetService(name)
	if err != nil {
		return err
	}

	if svc.Status == "stopped" {
		return fmt.Errorf("service '%s' is already stopped", name)
	}

	provider, ok := m.providers[svc.Type]
	if !ok {
		return fmt.Errorf("unsupported service type: %s", svc.Type)
	}

	if err := provider.Stop(ctx, svc); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	svc.Status = "stopped"
	svc.UpdatedAt = time.Now().Format(time.RFC3339)
	return m.state.SaveService(svc)
}

// Restart restarts a service
func (m *Manager) Restart(ctx context.Context, name string) error {
	svc, err := m.state.GetService(name)
	if err != nil {
		return err
	}

	provider, ok := m.providers[svc.Type]
	if !ok {
		return fmt.Errorf("unsupported service type: %s", svc.Type)
	}

	// Stop if running
	if svc.Status == "running" {
		if err := provider.Stop(ctx, svc); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	// Start
	if err := provider.Start(ctx, svc); err != nil {
		svc.Status = "failed"
		m.state.SaveService(svc)
		return fmt.Errorf("failed to start service: %w", err)
	}

	svc.Status = "running"
	svc.UpdatedAt = time.Now().Format(time.RFC3339)
	return m.state.SaveService(svc)
}

// Delete deletes a service
func (m *Manager) Delete(ctx context.Context, name string) error {
	svc, err := m.state.GetService(name)
	if err != nil {
		return err
	}

	provider, ok := m.providers[svc.Type]
	if !ok {
		return fmt.Errorf("unsupported service type: %s", svc.Type)
	}

	// Stop if running
	if svc.Status == "running" {
		if err := provider.Stop(ctx, svc); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	// Delete
	if err := provider.Delete(ctx, svc); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Remove from state
	return m.state.DeleteService(name)
}

// HealthCheck checks service health
func (m *Manager) HealthCheck(ctx context.Context, name string) (bool, error) {
	svc, err := m.state.GetService(name)
	if err != nil {
		return false, err
	}

	if svc.Status != "running" {
		return false, nil
	}

	provider, ok := m.providers[svc.Type]
	if !ok {
		return false, fmt.Errorf("unsupported service type: %s", svc.Type)
	}

	return provider.HealthCheck(ctx, svc)
}
