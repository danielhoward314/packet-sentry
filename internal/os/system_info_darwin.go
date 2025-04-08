//go:build darwin

package os

import (
	"context"
	"log/slog"
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

func (dsi *darwinSystemInfo) GetUniqueSystemIdentifier() (string, error) {
	return "", nil
}
