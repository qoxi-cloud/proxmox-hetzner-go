// Package main is the entry point for the pve-install CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/qoxi-cloud/proxmox-hetzner-go/pkg/version"
)

var (
	cfgFile    string
	saveConfig string
	verbose    bool
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "pve-install",
	Short: "TUI-based installer for Proxmox VE on Hetzner dedicated servers",
	Long: `pve-install is a TUI-based installer for Proxmox VE on Hetzner dedicated servers.

It provides a guided installation experience with:
- Hardware auto-detection
- Network configuration (NAT/external/both)
- SSH hardening
- Tailscale integration
- ZFS optimization`,
	Run: func(_ *cobra.Command, _ []string) {
		// TODO: Launch TUI here
		fmt.Println("Starting Proxmox VE installer TUI...")
		fmt.Println("TUI not implemented yet. Use 'pve-install --help' for available options.")
	},
}

// versionCmd shows version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, _ []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "pve-install %s\n", version.Full())
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: $HOME/.pve-install.yaml)")
	rootCmd.PersistentFlags().StringVarP(&saveConfig, "save-config", "s", "", "save configuration to file after input")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Bind flags to viper
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("save-config", rootCmd.PersistentFlags().Lookup("save-config"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not find home directory:", err)
			return
		}

		// Search config in home directory with name ".pve-install" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".pve-install")
	}

	// Read environment variables with PVE_ prefix
	viper.SetEnvPrefix("PVE")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

func main() {
	Execute()
}
