package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anisimov-anthony/vibector/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate sample configuration file",
	Long:  `Generate a sample .vibector.yaml configuration file in the current directory.`,
	RunE:  runConfigInit,
}

func init() {
	configCmd.AddCommand(configInitCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := ".vibector.yaml"

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	if err := config.GenerateSampleConfig(configPath); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Printf("Sample configuration file created: %s\n", configPath)
	fmt.Println("Edit this file to configure your thresholds.")

	return nil
}
