package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/carb1d3/cloudless/internal/container"
	"github.com/carb1d3/cloudless/internal/service"
	"github.com/carb1d3/cloudless/internal/state"
)

// Daemon represents the main Cloudless daemon
type Daemon struct {
	state          *state.State
	serviceManager *service.Manager
	containerMgr   *container.Manager
	ctx            context.Context
	cancel         context.CancelFunc
	restartCooldown map[string]time.Time
}

// New creates a new daemon instance
func New() (*Daemon, error) {
	ctx, cancel := context.WithCancel(context.Background())

	homeDir, err := os.UserHomeDir()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(homeDir, ".cloudless")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	st, err := state.New(stateDir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize state: %w", err)
	}

	containerMgr, err := container.NewManager()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize container manager: %w", err)
	}

	serviceMgr := service.NewManager(st, containerMgr)

	return &Daemon{
		state:           st,
		serviceManager:  serviceMgr,
		containerMgr:    containerMgr,
		ctx:             ctx,
		cancel:          cancel,
		restartCooldown: make(map[string]time.Time),
	}, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	fmt.Println("Starting Cloudless daemon...")

	// Load existing services and start them
	services, err := d.state.ListServices()
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	for _, svc := range services {
		if svc.AutoStart {
			fmt.Printf("Auto-starting service: %s\n", svc.Name)
			if err := d.serviceManager.Start(d.ctx, svc.Name); err != nil {
				fmt.Printf("Warning: failed to start service %s: %v\n", svc.Name, err)
			}
		}
	}

	// Create PID file
	homeDir, _ := os.UserHomeDir()
	pidFile := filepath.Join(homeDir, ".cloudless", "daemon.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer os.Remove(pidFile)

	fmt.Println("Cloudless daemon started successfully")
	fmt.Println("Press Ctrl+C to stop")

	// Start health check loop
	go d.healthCheckLoop()

	// Wait for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nShutting down daemon...")
	case <-d.ctx.Done():
	}

	return d.shutdown()
}

// healthCheckLoop periodically checks service health
func (d *Daemon) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			services, err := d.state.ListServices()
			if err != nil {
				continue
			}

			for _, svc := range services {
				if svc.Status == "running" {
					healthy, err := d.serviceManager.HealthCheck(d.ctx, svc.Name)
					if err != nil || !healthy {
						// Check if service is in cooldown period
						if lastRestart, exists := d.restartCooldown[svc.Name]; exists {
							if time.Since(lastRestart) < 2*time.Minute {
								fmt.Printf("Service %s is in cooldown, skipping restart\n", svc.Name)
								continue
							}
						}

						fmt.Printf("Service %s is unhealthy, attempting restart...\n", svc.Name)
						if err := d.serviceManager.Restart(d.ctx, svc.Name); err != nil {
							fmt.Printf("Failed to restart service %s: %v\n", svc.Name, err)
						} else {
							d.restartCooldown[svc.Name] = time.Now()
						}
					}
				}
			}
		case <-d.ctx.Done():
			return
		}
	}
}

// shutdown gracefully shuts down the daemon
func (d *Daemon) shutdown() error {
	d.cancel()

	// Stop all running services
	services, _ := d.state.ListServices()
	for _, svc := range services {
		if svc.Status == "running" {
			fmt.Printf("Stopping service: %s\n", svc.Name)
			d.serviceManager.Stop(context.Background(), svc.Name)
		}
	}

	fmt.Println("Daemon stopped")
	return nil
}

// IsRunning checks if the daemon is running
func IsRunning() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	pidFile := filepath.Join(homeDir, ".cloudless", "daemon.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil, nil
}

// Stop stops the daemon
func Stop() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pidFile := filepath.Join(homeDir, ".cloudless", "daemon.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daemon is not running")
		}
		return err
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	fmt.Println("Daemon stopped")
	return nil
}
