//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const installPath = `C:\Program Files\PacketSentry\`
const jsonFile = installPath + `\agentBootstrap.json`

// WriteInstallKey writes the install key to a JSON file.
func WriteInstallKey(installKey string) error {
	// Ensure the directory exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		fmt.Println("creating install directory", installPath)
		if err := os.MkdirAll(installPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Create JSON structure
	data := map[string]string{
		"installKey": installKey,
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Write JSON to file
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Println("wrote to file", jsonFile)
	return nil
}

// LockFileACLs sets the ACLs on the JSON file.
func LockFileACLs() error {
	// Command to lock ACLs using icacls
	cmd := exec.Command("icacls", jsonFile, "/inheritance:r", "/grant", "SYSTEM:F", "/grant", "Administrators:F", "/Q")

	// Capture output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to lock ACLs: %v\n%s", err, output)
	}

	fmt.Println("ACLs locked on file:", jsonFile)
	return nil
}

// CLI entry point
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  write-install-key <installKey>   - Writes the install key to a JSON file")
		fmt.Println("  lock-file-acls                  - Lock the file ACLs")
		os.Exit(1)
	}

	command := os.Args[1]

	switch strings.ToLower(command) {
	case "write-install-key":
		if len(os.Args) < 3 {
			fmt.Println("Usage: install_actions.exe write-install-key <installKey>")
			os.Exit(1)
		}
		installKey := os.Args[2]
		if err := WriteInstallKey(installKey); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	case "lock-file-acls":
		if err := LockFileACLs(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Usage: install_actions.exe <command> [args]")
		os.Exit(1)
	}
}
