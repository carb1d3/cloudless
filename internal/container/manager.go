package container

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Manager manages container operations
type Manager struct {
	client *client.Client
}

// NewManager creates a new container manager
func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify Docker is available
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("Docker is not available. Please ensure Docker is installed and running: %w", err)
	}

	return &Manager{client: cli}, nil
}

// ContainerConfig represents container configuration
type ContainerConfig struct {
	Image       string
	Name        string
	Env         []string
	Ports       map[string]string // containerPort -> hostPort
	Volumes     map[string]string // hostPath -> containerPath
	AutoRemove  bool
	NetworkMode string
}

// Create creates and starts a container
func (m *Manager) Create(ctx context.Context, config ContainerConfig) (string, error) {
	// Pull image if not exists
	if err := m.ensureImage(ctx, config.Image); err != nil {
		return "", fmt.Errorf("failed to ensure image: %w", err)
	}

	// Prepare port bindings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for containerPort, hostPort := range config.Ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid port %s: %w", containerPort, err)
		}

		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// Prepare volume bindings
	binds := []string{}
	for hostPath, containerPath := range config.Volumes {
		// Create host directory if it doesn't exist
		if err := os.MkdirAll(hostPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create volume directory: %w", err)
		}
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// Create container
	containerConfig := &container.Config{
		Image:        config.Image,
		Env:          config.Env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		AutoRemove:   config.AutoRemove,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	if config.NetworkMode != "" {
		hostConfig.NetworkMode = container.NetworkMode(config.NetworkMode)
	}

	resp, err := m.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		config.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

// Stop stops a container
func (m *Manager) Stop(ctx context.Context, containerID string) error {
	timeout := 10
	return m.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

// Remove removes a container
func (m *Manager) Remove(ctx context.Context, containerID string) error {
	return m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

// IsRunning checks if a container is running
func (m *Manager) IsRunning(ctx context.Context, containerID string) (bool, error) {
	info, err := m.client.ContainerInspect(ctx, containerID)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return info.State.Running, nil
}

// ensureImage pulls an image if it doesn't exist locally
func (m *Manager) ensureImage(ctx context.Context, imageName string) error {
	// Check if image exists
	_, _, err := m.client.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		return nil // Image exists
	}

	// Pull image
	fmt.Printf("Pulling image %s...\n", imageName)
	reader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Wait for pull to complete
	_, err = io.Copy(io.Discard, reader)
	return err
}

// Close closes the Docker client
func (m *Manager) Close() error {
	return m.client.Close()
}
