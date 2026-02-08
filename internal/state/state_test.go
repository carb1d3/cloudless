package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewState(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	if state == nil {
		t.Fatal("State is nil")
	}

	// Check that directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Fatal("State directory was not created")
	}
}

func TestSaveAndGetService(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	svc := &Service{
		Name:   "test-service",
		Type:   "postgres",
		Status: "created",
		Port:   5432,
		Config: map[string]string{
			"user":     "testuser",
			"password": "testpass",
		},
		AutoStart: true,
	}

	// Save service
	err = state.SaveService(svc)
	if err != nil {
		t.Fatalf("Failed to save service: %v", err)
	}

	// Get service
	retrieved, err := state.GetService("test-service")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	if retrieved.Name != svc.Name {
		t.Errorf("Expected name %s, got %s", svc.Name, retrieved.Name)
	}
	if retrieved.Type != svc.Type {
		t.Errorf("Expected type %s, got %s", svc.Type, retrieved.Type)
	}
	if retrieved.Port != svc.Port {
		t.Errorf("Expected port %d, got %d", svc.Port, retrieved.Port)
	}
}

func TestListServices(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Add multiple services
	services := []*Service{
		{Name: "service1", Type: "postgres", Status: "running"},
		{Name: "service2", Type: "postgres", Status: "stopped"},
		{Name: "service3", Type: "postgres", Status: "created"},
	}

	for _, svc := range services {
		if err := state.SaveService(svc); err != nil {
			t.Fatalf("Failed to save service: %v", err)
		}
	}

	// List services
	list, err := state.ListServices()
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}

	if len(list) != len(services) {
		t.Errorf("Expected %d services, got %d", len(services), len(list))
	}
}

func TestDeleteService(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	svc := &Service{
		Name:   "test-service",
		Type:   "postgres",
		Status: "created",
	}

	// Save service
	err = state.SaveService(svc)
	if err != nil {
		t.Fatalf("Failed to save service: %v", err)
	}

	// Delete service
	err = state.DeleteService("test-service")
	if err != nil {
		t.Fatalf("Failed to delete service: %v", err)
	}

	// Try to get deleted service
	_, err = state.GetService("test-service")
	if err == nil {
		t.Error("Expected error when getting deleted service")
	}
}

func TestUpdateServiceStatus(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	svc := &Service{
		Name:   "test-service",
		Type:   "postgres",
		Status: "created",
	}

	// Save service
	err = state.SaveService(svc)
	if err != nil {
		t.Fatalf("Failed to save service: %v", err)
	}

	// Update status
	err = state.UpdateServiceStatus("test-service", "running")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Get service and check status
	retrieved, err := state.GetService("test-service")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	if retrieved.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", retrieved.Status)
	}
}

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state and save a service
	state1, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	svc := &Service{
		Name:   "test-service",
		Type:   "postgres",
		Status: "running",
		Port:   5432,
	}

	err = state1.SaveService(svc)
	if err != nil {
		t.Fatalf("Failed to save service: %v", err)
	}

	// Create a new state instance (simulating daemon restart)
	state2, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second state: %v", err)
	}

	// Check that service was loaded
	retrieved, err := state2.GetService("test-service")
	if err != nil {
		t.Fatalf("Failed to get service from loaded state: %v", err)
	}

	if retrieved.Name != svc.Name {
		t.Errorf("Expected name %s, got %s", svc.Name, retrieved.Name)
	}

	// Check that state.json file exists
	stateFile := filepath.Join(tmpDir, "state.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("state.json file was not created")
	}
}
