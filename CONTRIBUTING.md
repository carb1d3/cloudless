# Contributing to Cloudless

Thank you for your interest in contributing to Cloudless! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker installed and running
- Basic understanding of Go and Docker

### Setting Up Development Environment

1. Clone the repository:
```bash
git clone https://github.com/carb1d3/cloudless.git
cd cloudless
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build -o cloudless ./cmd/cloudless
```

4. Run tests:
```bash
go test ./...
./test.sh
```

## Project Structure

```
cloudless/
├── cmd/
│   └── cloudless/          # CLI entry point
│       └── main.go        # Command definitions and CLI setup
├── internal/
│   ├── daemon/            # Daemon implementation
│   │   ├── daemon.go     # Main daemon logic, health checks
│   │   └── client.go     # Client-side service operations
│   ├── service/           # Service management
│   │   ├── manager.go    # Service lifecycle manager
│   │   └── postgres.go   # PostgreSQL provider implementation
│   ├── container/         # Container operations
│   │   └── manager.go    # Docker client wrapper
│   └── state/             # State persistence
│       ├── state.go      # State management logic
│       └── state_test.go # Unit tests
├── test.sh                # Integration test script
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── README.md              # User documentation
```

## Code Guidelines

### Go Style

- Follow standard Go conventions and formatting (use `gofmt`)
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and reasonably sized

### Error Handling

- Always handle errors explicitly
- Use `fmt.Errorf` with `%w` for error wrapping
- Provide context in error messages
- Don't silently ignore errors

### Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for meaningful test coverage
- Test error paths, not just happy paths

## Adding a New Service Provider

To add support for a new service type (e.g., Redis, MySQL):

1. Create a new file in `internal/service/` (e.g., `redis.go`)

2. Implement the `Provider` interface:
```go
type Provider interface {
    Create(ctx context.Context, name string) (*state.Service, error)
    Start(ctx context.Context, svc *state.Service) error
    Stop(ctx context.Context, svc *state.Service) error
    HealthCheck(ctx context.Context, svc *state.Service) (bool, error)
    Delete(ctx context.Context, svc *state.Service) error
}
```

3. Register your provider in `internal/service/manager.go`:
```go
func NewManager(st *state.State, containerMgr *container.Manager) *Manager {
    m := &Manager{
        state:        st,
        containerMgr: containerMgr,
        providers:    make(map[string]Provider),
    }
    
    m.providers["postgres"] = NewPostgresProvider(containerMgr)
    m.providers["redis"] = NewRedisProvider(containerMgr)  // Add your provider
    
    return m
}
```

4. Add tests for your provider

5. Update documentation

## Pull Request Process

1. Fork the repository and create a new branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes:
   - Write clean, well-tested code
   - Follow the code guidelines
   - Update documentation if needed

3. Run tests and linting:
```bash
go test ./...
go vet ./...
./test.sh
```

4. Commit your changes:
```bash
git add .
git commit -m "Brief description of changes"
```

5. Push to your fork:
```bash
git push origin feature/your-feature-name
```

6. Create a Pull Request:
   - Provide a clear description of the changes
   - Link any related issues
   - Ensure CI passes

## Development Tips

### Running the Daemon Locally

For development, you can run the daemon in the foreground:
```bash
go run ./cmd/cloudless daemon start
```

### Testing with Docker

Ensure Docker is running before testing container operations:
```bash
docker ps
```

### Debugging

You can add debug logging throughout the code. Consider using structured logging for production code.

### State Management

The daemon stores state in `~/.cloudless/`. During development, you can:
- Check state: `cat ~/.cloudless/state.json`
- Clear state: `rm -rf ~/.cloudless/`

## Reporting Issues

When reporting issues, please include:
- Go version (`go version`)
- Docker version (`docker version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## Feature Requests

We welcome feature requests! Please:
- Check if a similar request exists
- Provide clear use cases
- Explain why it would benefit users

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help maintain a positive community

## Questions?

Feel free to:
- Open an issue for questions
- Start a discussion
- Reach out to maintainers

Thank you for contributing to Cloudless!
