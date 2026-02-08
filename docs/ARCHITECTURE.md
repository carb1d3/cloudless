# Cloudless Architecture

## Overview

Cloudless is designed as a local control plane that manages containerized services. It follows a daemon-based architecture with clear separation of concerns.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       CLI (cloudless)                        │
│  Commands: daemon start|stop|status, service create|list... │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Daemon Process                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │            Daemon Controller                          │  │
│  │  - Process management                                 │  │
│  │  - Signal handling (SIGINT, SIGTERM)                  │  │
│  │  - Health check loop (30s interval)                   │  │
│  │  - Auto-restart with cooldown (2min)                  │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                        │
│                     ▼                                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │            Service Manager                            │  │
│  │  - Create, Start, Stop, Restart, Delete              │  │
│  │  - Provider registry and delegation                   │  │
│  │  - Service health checks                              │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                        │
│         ┌───────────┴────────────┐                          │
│         ▼                        ▼                          │
│  ┌─────────────┐         ┌─────────────┐                   │
│  │  Postgres   │         │   Future    │                   │
│  │  Provider   │         │  Providers  │                   │
│  │             │         │  (Redis,    │                   │
│  │             │         │   MySQL,    │                   │
│  │             │         │   etc.)     │                   │
│  └──────┬──────┘         └─────────────┘                   │
│         │                                                    │
│         ▼                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Container Manager                           │   │
│  │  - Docker client wrapper                             │   │
│  │  - Image pull and management                         │   │
│  │  - Container lifecycle (create, start, stop, remove) │   │
│  │  - Port and volume management                        │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                        │
│                     ▼                                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │             State Manager                            │   │
│  │  - Service configuration persistence                 │   │
│  │  - JSON-based storage (~/.cloudless/state.json)     │   │
│  │  - Thread-safe access with mutex                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      Docker Engine                           │
│  - Container runtime                                         │
│  - Image registry                                            │
│  - Network and volume management                             │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Local Filesystem                           │
│  ~/.cloudless/                                              │
│  ├── state.json          # Service configuration            │
│  ├── daemon.pid          # Daemon process ID                │
│  └── data/                                                   │
│      └── <service>/      # Per-service persistent data      │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. CLI (cmd/cloudless)

The command-line interface built with Cobra. Provides two main command groups:

- **daemon commands**: Control the daemon lifecycle
  - `start`: Launch the daemon process
  - `stop`: Gracefully shutdown the daemon
  - `status`: Check if daemon is running

- **service commands**: Manage services
  - `create <type> <name>`: Create and start a new service
  - `list`: List all services
  - `start <name>`: Start a stopped service
  - `stop <name>`: Stop a running service
  - `restart <name>`: Restart a service
  - `delete <name>`: Remove a service completely

### 2. Daemon (internal/daemon)

The main daemon process that runs continuously in the background.

**Key responsibilities:**
- Initialize and manage the service manager
- Run health check loop every 30 seconds
- Auto-restart unhealthy services with cooldown
- Handle graceful shutdown on signals
- Manage PID file for process tracking

**Health Check Logic:**
```
Every 30 seconds:
  For each running service:
    Check health
    If unhealthy:
      Check cooldown (2 minutes)
      If not in cooldown:
        Attempt restart
        Record restart time
```

### 3. Service Manager (internal/service)

Orchestrates service lifecycle through provider implementations.

**Provider Interface:**
```go
type Provider interface {
    Create(ctx context.Context, name string) (*state.Service, error)
    Start(ctx context.Context, svc *state.Service) error
    Stop(ctx context.Context, svc *state.Service) error
    HealthCheck(ctx context.Context, svc *state.Service) (bool, error)
    Delete(ctx context.Context, svc *state.Service) error
}
```

**Responsibilities:**
- Delegate operations to appropriate providers
- Update service state
- Coordinate with state manager
- Handle provider registration

### 4. PostgreSQL Provider (internal/service/postgres.go)

Implements the Provider interface for PostgreSQL.

**Create:**
- Find available port
- Generate default credentials
- Create service configuration

**Start:**
- Prepare data directory
- Create Docker container with:
  - postgres:16-alpine image
  - Environment variables for credentials
  - Port mapping
  - Volume for data persistence
  - Restart policy: unless-stopped
- Wait for PostgreSQL to be ready

**Health Check:**
- Verify container is running
- Attempt database connection
- Execute ping query

**Stop:**
- Stop Docker container gracefully

**Delete:**
- Remove container
- Delete data directory

### 5. Container Manager (internal/container)

Abstracts Docker operations.

**Key features:**
- Docker client initialization and health check
- Image pulling with progress
- Container creation with configuration
- Port binding and volume mounting
- Container lifecycle management
- Status checking

### 6. State Manager (internal/state)

Manages service configuration persistence.

**Data Structure:**
```go
type Service struct {
    Name        string            // Unique service name
    Type        string            // Service type (postgres, etc.)
    Status      string            // created, running, stopped, failed
    ContainerID string            // Docker container ID
    Port        int               // Host port
    Config      map[string]string // Service-specific configuration
    AutoStart   bool              // Start on daemon launch
    CreatedAt   string            // ISO 8601 timestamp
    UpdatedAt   string            // ISO 8601 timestamp
}
```

**Storage:**
- JSON file: `~/.cloudless/state.json`
- Thread-safe with RWMutex
- Automatic persistence on changes
- Loaded on daemon start

## Data Flow

### Creating a Service

1. User runs: `cloudless service create postgres mydb`
2. CLI checks if daemon is running
3. CLI creates service manager and calls Create
4. Service manager:
   - Calls PostgreSQL provider Create
   - Saves service to state
   - Calls provider Start
5. PostgreSQL provider:
   - Creates data directory
   - Calls container manager to create container
   - Waits for database to be ready
6. Container manager:
   - Pulls image if needed
   - Creates and starts container
   - Returns container ID
7. Service manager updates state to "running"
8. CLI displays connection information

### Health Check Cycle

1. Daemon health check loop triggers (every 30s)
2. For each running service:
   - Service manager calls provider HealthCheck
   - PostgreSQL provider:
     - Checks container status
     - Attempts database connection
     - Returns health status
   - If unhealthy and not in cooldown:
     - Service manager calls Restart
     - Cooldown period set (2 minutes)

### Daemon Shutdown

1. User presses Ctrl+C or runs `cloudless daemon stop`
2. Daemon receives SIGINT or SIGTERM
3. Daemon shutdown sequence:
   - Load all services from state
   - For each running service:
     - Stop container gracefully
   - Remove PID file
   - Exit

## File System Layout

```
~/.cloudless/
├── state.json           # Service configurations (JSON)
├── daemon.pid           # Daemon process ID
└── data/
    ├── mydb/            # PostgreSQL data for service "mydb"
    │   └── ...
    └── other-service/   # Data for another service
        └── ...
```

## Design Decisions

### Why JSON for State?

- Human-readable for debugging
- Simple to implement
- Sufficient for local development use case
- Easy migration path to database if needed

### Why Separate Providers?

- Extensibility: Easy to add new service types
- Separation of concerns: Each service has unique requirements
- Testability: Can test providers independently
- Maintainability: Clear boundaries

### Why Docker?

- Industry standard for containerization
- Excellent isolation
- Built-in image registry
- Handles networking and volumes
- Restart policies

### Why Background Daemon?

- Continuous health monitoring
- Auto-recovery without user intervention
- Service persistence across reboots (with auto-start)
- Single source of truth for service state

## Future Enhancements

### Planned Features

1. **Additional Service Providers**
   - Redis
   - MySQL
   - MongoDB
   - RabbitMQ

2. **Service Discovery**
   - DNS-based service discovery
   - Environment variable injection
   - Service-to-service networking

3. **Configuration Management**
   - Custom service configurations
   - Template-based provisioning
   - Configuration validation

4. **Backup and Restore**
   - Automated backups
   - Point-in-time recovery
   - Export/import functionality

5. **Resource Management**
   - CPU and memory limits
   - Storage quotas
   - Resource monitoring

6. **Web UI**
   - Service dashboard
   - Real-time logs
   - Performance metrics

### Potential Improvements

- Replace JSON state with embedded database (SQLite)
- Add structured logging (zerolog, zap)
- Implement gRPC API for better client/daemon communication
- Add Prometheus metrics export
- Support for custom Docker networks
- Multi-host support via Docker contexts

## Security Considerations

### Current Security Features

- Services run in isolated containers
- Default credentials (should be changed in production use)
- Local-only by default (no external exposure)
- PID file prevents multiple daemon instances

### Security TODOs

- Credential encryption in state file
- TLS for inter-service communication
- User authentication for CLI
- Audit logging
- Network policies
- Secret management integration

## Testing Strategy

### Unit Tests

- State management operations
- Service provider logic
- Container manager operations

### Integration Tests

- CLI command validation
- End-to-end service creation
- Daemon lifecycle management

### Manual Tests

- Full daemon + service workflow
- Health check and recovery
- Multi-service scenarios
- Error conditions and edge cases

## Performance Characteristics

### Resource Usage

- Daemon: ~10-20 MB memory
- PostgreSQL container: ~100-200 MB memory
- Disk: Depends on service data

### Scalability

- Designed for local development
- Can manage 10-20 services comfortably
- Health check interval: 30 seconds
- Restart cooldown: 2 minutes

### Bottlenecks

- Docker daemon performance
- Local disk I/O for databases
- Single-threaded health checks (can be parallelized)
