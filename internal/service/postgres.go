package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/carb1d3/cloudless/internal/container"
	"github.com/carb1d3/cloudless/internal/state"
	_ "github.com/lib/pq"
)

// PostgresProvider implements the Provider interface for PostgreSQL
type PostgresProvider struct {
	containerMgr *container.Manager
}

// NewPostgresProvider creates a new PostgreSQL provider
func NewPostgresProvider(containerMgr *container.Manager) *PostgresProvider {
	return &PostgresProvider{
		containerMgr: containerMgr,
	}
}

// Create creates a new PostgreSQL service
func (p *PostgresProvider) Create(ctx context.Context, name string) (*state.Service, error) {
	// Generate default configuration
	config := map[string]string{
		"user":     "cloudless",
		"password": "cloudless",
		"database": "cloudless",
	}

	// Find available port
	port, err := p.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	svc := &state.Service{
		Name:      name,
		Type:      "postgres",
		Status:    "created",
		Port:      port,
		Config:    config,
		AutoStart: true,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	return svc, nil
}

// Start starts the PostgreSQL service
func (p *PostgresProvider) Start(ctx context.Context, svc *state.Service) error {
	// Prepare data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dataDir := filepath.Join(homeDir, ".cloudless", "data", svc.Name)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Prepare environment variables
	env := []string{
		fmt.Sprintf("POSTGRES_USER=%s", svc.Config["user"]),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", svc.Config["password"]),
		fmt.Sprintf("POSTGRES_DB=%s", svc.Config["database"]),
	}

	// Container configuration
	containerConfig := container.ContainerConfig{
		Image: "postgres:16-alpine",
		Name:  fmt.Sprintf("cloudless-%s", svc.Name),
		Env:   env,
		Ports: map[string]string{
			"5432": fmt.Sprintf("%d", svc.Port),
		},
		Volumes: map[string]string{
			dataDir: "/var/lib/postgresql/data",
		},
		AutoRemove: false,
	}

	// Create and start container
	containerID, err := p.containerMgr.Create(ctx, containerConfig)
	if err != nil {
		return err
	}

	svc.ContainerID = containerID

	// Wait for PostgreSQL to be ready
	if err := p.waitForReady(ctx, svc); err != nil {
		return fmt.Errorf("PostgreSQL failed to start: %w", err)
	}

	return nil
}

// Stop stops the PostgreSQL service
func (p *PostgresProvider) Stop(ctx context.Context, svc *state.Service) error {
	if svc.ContainerID == "" {
		return nil
	}

	return p.containerMgr.Stop(ctx, svc.ContainerID)
}

// Delete deletes the PostgreSQL service
func (p *PostgresProvider) Delete(ctx context.Context, svc *state.Service) error {
	if svc.ContainerID != "" {
		if err := p.containerMgr.Remove(ctx, svc.ContainerID); err != nil {
			return err
		}
	}

	// Remove data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".cloudless", "data", svc.Name)
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("failed to remove data directory: %w", err)
	}

	return nil
}

// HealthCheck checks if PostgreSQL is healthy
func (p *PostgresProvider) HealthCheck(ctx context.Context, svc *state.Service) (bool, error) {
	if svc.ContainerID == "" {
		return false, nil
	}

	// Check if container is running
	running, err := p.containerMgr.IsRunning(ctx, svc.ContainerID)
	if err != nil || !running {
		return false, err
	}

	// Try to connect to PostgreSQL
	connStr := fmt.Sprintf("host=localhost port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		svc.Port, svc.Config["user"], svc.Config["password"], svc.Config["database"])

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return false, nil
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false, nil
	}

	return true, nil
}

// waitForReady waits for PostgreSQL to be ready
func (p *PostgresProvider) waitForReady(ctx context.Context, svc *state.Service) error {
	connStr := fmt.Sprintf("host=localhost port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=2",
		svc.Port, svc.Config["user"], svc.Config["password"], svc.Config["database"])

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for PostgreSQL to be ready")
		case <-ticker.C:
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				continue
			}

			if err := db.Ping(); err != nil {
				db.Close()
				continue
			}

			db.Close()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// findAvailablePort finds an available port on the host
func (p *PostgresProvider) findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
