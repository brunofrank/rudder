package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type RudderConfig struct {
	Rudder struct {
		DefaultService string `yaml:"default_service"`
		Commands       map[string]interface{} `yaml:"commands"`
	} `yaml:"rudder"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rudder <command> [args...]")
		os.Exit(1)
	}

	// Read and parse rudder.yml
	configFile, err := os.ReadFile(".rudder.yml")
	if err != nil {
		fmt.Printf("Error reading .rudder.yml: %v\n", err)
		os.Exit(1)
	}

	var config RudderConfig
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		fmt.Printf("Error parsing .rudder.yml: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Get the command definition from config
	cmdDef, exists := config.Rudder.Commands[command]
	if !exists {
		fmt.Printf("Command '%s' not found in .rudder.yml\n", command)
		os.Exit(1)
	}

	// Handle different types of command definitions
	switch cmd := cmdDef.(type) {
	case string:
		executeCommand(cmd, args, config.Rudder.DefaultService)
	case []interface{}:
		for _, cmdStr := range cmd {
			if str, ok := cmdStr.(string); ok {
				executeCommand(str, args, config.Rudder.DefaultService)
			}
		}
	default:
		fmt.Printf("Invalid command definition for '%s'\n", command)
		os.Exit(1)
	}
}

func executeCommand(cmdStr string, args []string, defaultService string) {
	// Check if command contains @ for specific service
	parts := strings.Split(cmdStr, "@")
	cmd := parts[0]
	service := defaultService

	if len(parts) > 1 {
		service = strings.TrimSpace(parts[1])
	}

  cmd = strings.ReplaceAll(cmd, "$ARGS", strings.Join(args, " "))

	// Prepare the command
	var finalCmd *exec.Cmd

  fmt.Println("Executing command locally:", cmd)

	if service == "host" {
		// Local execution
		finalCmd = exec.Command("bash", "-c", cmd)
	} else {
		// Docker Compose execution
		// Load environment variables
		finalCmd = exec.Command("bash", "-c", fmt.Sprintf("docker compose run --rm %s %s", service, cmd))
	}

	// Set up pipes for output
	finalCmd.Stdout = os.Stdout
	finalCmd.Stderr = os.Stderr
	finalCmd.Stdin = os.Stdin

	// Execute the command
	if err := finalCmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
