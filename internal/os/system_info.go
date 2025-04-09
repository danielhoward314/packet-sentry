package os

import (
	"context"
	"log/slog"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const (
	logAttrValSvcName = "systemInfo"
)

// SystemInfo is the interface for platform-specific operations for getting info about the system
type SystemInfo interface {
	GetUniqueSystemIdentifier() (string, error)
}

// NewSystemInfo returns a platform-specific implementation of the SystemInfo interface
func NewSystemInfo(ctx context.Context, baseLogger *slog.Logger) SystemInfo {
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))
	return newSystemInfo(ctx, childLogger)
}
