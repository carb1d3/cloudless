package main

import (
	"fmt"
	"os"

	"github.com/carb1d3/cloudless/internal/daemon"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloudless",
	Short: "Cloud-like managed services, running locally",
	Long:  `Cloudless is a local control plane that brings cloud-style managed services to your own machine.`,
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the Cloudless daemon",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Cloudless daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := daemon.New()
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}
		return d.Start()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		running, err := daemon.IsRunning()
		if err != nil {
			return err
		}
		if running {
			fmt.Println("Daemon is running")
		} else {
			fmt.Println("Daemon is not running")
		}
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Cloudless daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Stop()
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
}

var createServiceCmd = &cobra.Command{
	Use:   "create [service-type] [name]",
	Short: "Create a new service",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceType := args[0]
		name := args[1]
		return daemon.CreateService(serviceType, name)
	},
}

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.ListServices()
	},
}

var deleteServiceCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.DeleteService(args[0])
	},
}

var startServiceCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.StartService(args[0])
	},
}

var stopServiceCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.StopService(args[0])
	},
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart [name]",
	Short: "Restart a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.RestartService(args[0])
	},
}

func init() {
	daemonCmd.AddCommand(startCmd)
	daemonCmd.AddCommand(statusCmd)
	daemonCmd.AddCommand(stopCmd)

	serviceCmd.AddCommand(createServiceCmd)
	serviceCmd.AddCommand(listServicesCmd)
	serviceCmd.AddCommand(deleteServiceCmd)
	serviceCmd.AddCommand(startServiceCmd)
	serviceCmd.AddCommand(stopServiceCmd)
	serviceCmd.AddCommand(restartServiceCmd)

	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(serviceCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
