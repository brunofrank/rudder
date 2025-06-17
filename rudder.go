package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type RudderConfig struct {
	Rudder struct {
		DefaultService string                 `yaml:"default_service"`
		Commands       map[string]interface{} `yaml:"commands"`
	} `yaml:"rudder"`
}

type Release struct {
	TagName string `json:"tag_name"`
}

func checkForUpdates() error {
	// Get current version
	currentVersion := "v0.0.0" // This should be replaced with actual version during build
	versionFile := filepath.Join(os.Getenv("HOME"), ".rudder", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	// Get latest version from GitHub
	resp, err := http.Get("https://api.github.com/repos/brunofrank/rudder/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	if release.TagName > currentVersion {
		fmt.Printf("New version available: %s (current: %s)\n", release.TagName, currentVersion)
		fmt.Println("Updating Rudder...")

		// Run the install script to update
		installScript := filepath.Join(os.Getenv("HOME"), ".rudder", "install.sh")
		if _, err := os.Stat(installScript); os.IsNotExist(err) {
			// Download install script if it doesn't exist
			resp, err := http.Get("https://raw.githubusercontent.com/bfscordeiro/rudder/main/install.sh")
			if err != nil {
				return fmt.Errorf("failed to download install script: %v", err)
			}
			defer resp.Body.Close()

			scriptData, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read install script: %v", err)
			}

			if err := os.WriteFile(installScript, scriptData, 0755); err != nil {
				return fmt.Errorf("failed to save install script: %v", err)
			}
		}

		// Execute the install script
		cmd := exec.Command(installScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update: %v", err)
		}

		// Save new version
		if err := os.WriteFile(versionFile, []byte(release.TagName), 0644); err != nil {
			return fmt.Errorf("failed to save version: %v", err)
		}

		fmt.Println("Update completed successfully!")
		return nil
	}

	fmt.Println("Rudder is up to date!")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rudder <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[1]

	if command == "init" {
		createRudderConfig()
		return
	}

	if command == "update" {
		if err := checkForUpdates(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
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

	args := os.Args[2:]

	// Get the command definition from config
	cmdDef, _ := config.Rudder.Commands[command]

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
		executeCommand(os.Args[1], args, config.Rudder.DefaultService)
	}
}

func createRudderConfig() {
	defaultConfig := `rudder:
  default_service: web
  commands:
    ssh: bash -l
    yarn: yarn $ARGS
    pristine:
      - echo "This will destroy your containers and replace them with new ones." @host
      # - docker compose down -v @host
      # - docker compose up --build --force-recreate --no-start @host
    setup:
      - echo "Setting up project..." @host
      # - yarn install
  `

	if _, err := os.Stat(".rudder.yml"); err == nil {
		fmt.Println("Error: .rudder.yml already exists")
		os.Exit(1)
	}

	if err := os.WriteFile(".rudder.yml", []byte(defaultConfig), 0644); err != nil {
		fmt.Printf("Error creating .rudder.yml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created .rudder.yml with default configuration")
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

	if service == "host" {
		// Local execution
		finalCmd = exec.Command("bash", "-c", cmd)
	} else {
		// Docker Compose execution
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
