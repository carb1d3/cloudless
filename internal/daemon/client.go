package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/carb1d3/cloudless/internal/container"
	"github.com/carb1d3/cloudless/internal/service"
	"github.com/carb1d3/cloudless/internal/state"
)

// CreateService creates a new service
func CreateService(serviceType, name string) error {
	running, err := IsRunning()
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("daemon is not running. Start it with 'cloudless daemon start'")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stateDir := filepath.Join(homeDir, ".cloudless")
	st, err := state.New(stateDir)
	if err != nil {
		return err
	}

	containerMgr, err := container.NewManager()
	if err != nil {
		return err
	}

	serviceMgr := service.NewManager(st, containerMgr)

	ctx := context.Background()
	if err := serviceMgr.Create(ctx, serviceType, name); err != nil {
		return err
	}

	fmt.Printf("Service '%s' created successfully\n", name)
	fmt.Printf("Type: %s\n", serviceType)

	// Get connection info
	svc, err := st.GetService(name)
	if err == nil && svc.Status == "running" {
		if svc.Type == "postgres" {
			fmt.Printf("\nConnection details:\n")
			fmt.Printf("  Host: localhost\n")
			fmt.Printf("  Port: %d\n", svc.Port)
			fmt.Printf("  Database: %s\n", svc.Config["database"])
			fmt.Printf("  User: %s\n", svc.Config["user"])
			fmt.Printf("  Password: %s\n", svc.Config["password"])
			fmt.Printf("\nConnection string:\n")
			fmt.Printf("  postgresql://%s:%s@localhost:%d/%s\n",
				svc.Config["user"], svc.Config["password"], svc.Port, svc.Config["database"])
		}
	}

	return nil
}

// ListServices lists all services
func ListServices() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stateDir := filepath.Join(homeDir, ".cloudless")
	st, err := state.New(stateDir)
	if err != nil {
		return err
	}

	services, err := st.ListServices()
	if err != nil {
		return err
	}

	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tPORT\tAUTO-START")
	fmt.Fprintln(w, "----\t----\t------\t----\t----------")

	for _, svc := range services {
		autoStart := "no"
		if svc.AutoStart {
			autoStart = "yes"
		}
		port := "-"
		if svc.Port > 0 {
			port = fmt.Sprintf("%d", svc.Port)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", svc.Name, svc.Type, svc.Status, port, autoStart)
	}

	w.Flush()
	return nil
}

// DeleteService deletes a service
func DeleteService(name string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stateDir := filepath.Join(homeDir, ".cloudless")
	st, err := state.New(stateDir)
	if err != nil {
		return err
	}

	containerMgr, err := container.NewManager()
	if err != nil {
		return err
	}

	serviceMgr := service.NewManager(st, containerMgr)

	ctx := context.Background()
	if err := serviceMgr.Delete(ctx, name); err != nil {
		return err
	}

	fmt.Printf("Service '%s' deleted successfully\n", name)
	return nil
}

// StartService starts a service
func StartService(name string) error {
	return executeServiceCommand(name, "start")
}

// StopService stops a service
func StopService(name string) error {
	return executeServiceCommand(name, "stop")
}

// RestartService restarts a service
func RestartService(name string) error {
	return executeServiceCommand(name, "restart")
}

func executeServiceCommand(name, command string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stateDir := filepath.Join(homeDir, ".cloudless")
	st, err := state.New(stateDir)
	if err != nil {
		return err
	}

	containerMgr, err := container.NewManager()
	if err != nil {
		return err
	}

	serviceMgr := service.NewManager(st, containerMgr)

	ctx := context.Background()

	var cmdErr error
	switch command {
	case "start":
		cmdErr = serviceMgr.Start(ctx, name)
	case "stop":
		cmdErr = serviceMgr.Stop(ctx, name)
	case "restart":
		cmdErr = serviceMgr.Restart(ctx, name)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	if cmdErr != nil {
		return cmdErr
	}

	// Use proper past tense for each command
	pastTense := map[string]string{
		"start":   "started",
		"stop":    "stopped",
		"restart": "restarted",
	}
	fmt.Printf("Service '%s' %s successfully\n", name, pastTense[command])
	return nil
}
