//go:build windows

package os

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type windowsSystemInfo struct {
	ctx    context.Context
	logger *slog.Logger
}

func newSystemInfo(ctx context.Context, logger *slog.Logger) SystemInfo {
	return &windowsSystemInfo{
		ctx:    ctx,
		logger: logger,
	}
}

// // GetUniqueSystemIdentifier is the windows implementation for getting a unique system identifier
func (wsi *windowsSystemInfo) GetUniqueSystemIdentifier() (string, error) {
	// https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-computersystemproduct
	// This value comes from the UUID member of the System Information structure in the SMBIOS information.
	cmd := exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run PowerShell command: %w", err)
	}

	uuid := strings.TrimSpace(string(output))

	if uuid == "" {
		return "", fmt.Errorf("UUID not found")
	}

	return uuid, nil
}
