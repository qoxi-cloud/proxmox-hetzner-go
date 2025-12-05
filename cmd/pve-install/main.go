// Package main is the entry point for the Proxmox VE installer.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information (set at build time).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Placeholder to ensure imports are used.
var (
	_ tea.Model
	_ spinner.Model
	_ = lipgloss.NewStyle
	_ = viper.New
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "pve-install",
		Short:   "TUI-based Proxmox VE installer for Hetzner dedicated servers",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Proxmox VE Installer")
			fmt.Println("Run with --help for usage information")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
