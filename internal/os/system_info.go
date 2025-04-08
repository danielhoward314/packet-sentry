package os

import (
	"context"
	"log/slog"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const (
	logAttrValSvcName = "systemInfo"
)

type SystemInfo interface {
	GetUniqueSystemIdentifier() (string, error)
}

func NewSystemInfo(ctx context.Context, baseLogger *slog.Logger) SystemInfo {
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))
	return newSystemInfo(ctx, childLogger)
}
