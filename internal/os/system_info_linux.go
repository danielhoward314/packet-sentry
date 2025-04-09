//go:build linux

package os

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
)

type linuxSystemInfo struct {
	ctx    context.Context
	logger *slog.Logger
}

func newSystemInfo(ctx context.Context, logger *slog.Logger) SystemInfo {
	return &linuxSystemInfo{
		ctx:    ctx,
		logger: logger,
	}
}

// GetUniqueSystemIdentifier is the linux implementation for getting a unique system identifier
func (lsi *linuxSystemInfo) GetUniqueSystemIdentifier() (string, error) {
	paths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id", // fallback
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			id := strings.TrimSpace(string(data))
			if id != "" {
				return id, nil
			}
		}
	}
	return "", errors.New("machine-id not found")
}
