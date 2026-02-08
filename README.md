# cloudless

Cloud-like managed services, running locally.

Cloudless is a local control plane that brings cloud-style managed services to your own machine. It runs as a background daemon that provisions and manages services like PostgreSQL using containers, handling lifecycle, health checks, restarts, and state. The goal is one-click local infrastructure with sensible defaults—no cloud account, no YAML, no internet, just reliable managed services running offline.

## Features

- **Background Daemon**: Runs as a persistent background process managing all your local services
- **Container-Based**: Uses Docker to run services in isolated, reproducible containers
- **Lifecycle Management**: Start, stop, restart services with simple commands
- **Health Checks**: Automatic health monitoring with auto-recovery
- **State Persistence**: Service configurations and state survive daemon restarts
- **Sensible Defaults**: Zero configuration required - services work out of the box
- **Offline First**: No internet connection required - everything runs locally

## Prerequisites

- Docker must be installed and running on your machine
- Go 1.21 or later (for building from source)

## Installation

### Build from Source

```bash
git clone https://github.com/carb1d3/cloudless.git
cd cloudless
go build -o cloudless ./cmd/cloudless
sudo mv cloudless /usr/local/bin/  # Optional: install globally
```

## Quick Start

### 1. Start the Daemon

```bash
cloudless daemon start
```

The daemon will run in the foreground. Press Ctrl+C to stop it, or run it in the background:

```bash
cloudless daemon start &
```

### 2. Create a PostgreSQL Service

```bash
cloudless service create postgres mydb
```

This will:
- Pull the PostgreSQL Docker image (if not already available)
- Create a new PostgreSQL instance with default credentials
- Start the service automatically
- Display connection details

### 3. List Services

```bash
cloudless service list
```

Output:
```
NAME    TYPE      STATUS    PORT    AUTO-START
----    ----      ------    ----    ----------
mydb    postgres  running   54321   yes
```

### 4. Manage Services

```bash
# Stop a service
cloudless service stop mydb

# Start a service
cloudless service start mydb

# Restart a service
cloudless service restart mydb

# Delete a service
cloudless service delete mydb
```

### 5. Check Daemon Status

```bash
cloudless daemon status
```

### 6. Stop the Daemon

```bash
cloudless daemon stop
```

## Services

### PostgreSQL

When you create a PostgreSQL service, Cloudless automatically:
- Assigns a random available port
- Creates default credentials (user: `cloudless`, password: `cloudless`, database: `cloudless`)
- Persists data in `~/.cloudless/data/<service-name>`
- Configures auto-restart on daemon start
- Sets up health checks

**Connection Example:**

```bash
psql -h localhost -p 54321 -U cloudless -d cloudless
# Password: cloudless
```

**Connection String:**

```
postgresql://cloudless:cloudless@localhost:54321/cloudless
```

## Architecture

Cloudless consists of several key components:

- **Daemon**: Main process that manages the lifecycle of all services
- **Service Manager**: Handles creation, lifecycle, and health checks for services
- **Container Manager**: Abstracts Docker operations for container management
- **State Manager**: Persists service configuration and state to disk
- **Service Providers**: Pluggable implementations for different service types (PostgreSQL, etc.)

All data is stored in `~/.cloudless/`:
- `state.json`: Service configurations and metadata
- `data/<service-name>`: Persistent data volumes for each service
- `daemon.pid`: Process ID of the running daemon

## Development

### Project Structure

```
cloudless/
├── cmd/
│   └── cloudless/          # CLI entry point
├── internal/
│   ├── daemon/            # Daemon implementation and client
│   ├── service/           # Service manager and providers
│   ├── container/         # Container management
│   └── state/             # State persistence
└── pkg/
    └── api/               # Public APIs (future)
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o cloudless ./cmd/cloudless
```

## Roadmap

- [ ] Additional service types (Redis, MySQL, MongoDB)
- [ ] Service discovery and networking between services
- [ ] Custom configurations per service
- [ ] Web UI for service management
- [ ] Backup and restore functionality
- [ ] Resource limits and quotas
- [ ] Service templates and presets

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

