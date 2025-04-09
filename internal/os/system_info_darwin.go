//go:build darwin

package os

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
)

type darwinSystemInfo struct {
	ctx    context.Context
	logger *slog.Logger
}

func newSystemInfo(ctx context.Context, logger *slog.Logger) SystemInfo {
	return &darwinSystemInfo{
		ctx:    ctx,
		logger: logger,
	}
}

// GetUniqueSystemIdentifier is the darwin implementation for getting a unique system identifier
func (dsi *darwinSystemInfo) GetUniqueSystemIdentifier() (string, error) {
	out, err := exec.Command("/usr/sbin/ioreg", "-l").Output()
	if err != nil {
		dsi.logger.Error("failed to get system identifier", "error", err)
		return "", err
	}
	serialNumber := ""
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, "IOPlatformSerialNumber") {
			s := strings.Split(l, " ")
			serialNumber = s[len(s)-1]
			serialNumber = strings.Trim(serialNumber, "\"")
			break
		}
	}
	return serialNumber, nil
}
