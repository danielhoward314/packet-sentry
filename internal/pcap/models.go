package pcap

import (
	"fmt"
	"log/slog"
	"time"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

// CaptureConfig holds configuration for a packet capture
type CaptureConfig struct {
	BPF         string        `json:"bpf"`
	DeviceName  string        `json:"deviceName"`
	Promiscuous bool          `json:"promiscuous"`
	SnapLen     int32         `json:"snapLen"`
	Timeout     time.Duration `json:"timeout"`
}

// LogValue implements the slog.LogValuer interface for the CaptureConfig struct
func (cc *CaptureConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String(psLog.KeyBPF, cc.BPF),
		slog.String(psLog.KeyDeviceName, cc.DeviceName),
		slog.Bool(psLog.KeyPromiscuous, cc.Promiscuous),
		slog.Int64(psLog.KeySnapLen, int64(cc.SnapLen)),
		slog.String(psLog.KeyTimeout, cc.Timeout.String()),
	)
}

type InterfaceDetails struct {
	Name string `json:"name"`
}

type ReportInterfacesRequest struct {
	Interfaces  []*InterfaceDetails `json:"interfaces"`
	PCapVersion string              `json:"pcapVersion"`
}

type BPFResponse struct {
	Create map[string]map[uint64]*CaptureConfig `json:"create"`
	Update map[string]map[uint64]*CaptureConfig `json:"update"`
	Delete map[string]map[uint64]*CaptureConfig `json:"delete"`
}

type CachedBPFConfig struct {
	InterfacesToBPFAssociations map[string]map[uint64]*CaptureConfig `json:"config"`
}

// PacketCaptureNotFound is a custom error type signaling that a packet capture is not found in the current map of BPF associations.
type PacketCaptureNotFound struct {
	FilterHash uint64
	Filter     string
}

// Error implements the Error interface for the PacketCaptureNotFound
func (e *PacketCaptureNotFound) Error() string {
	return fmt.Sprintf("packet capture not found by hash %d of BPF %s", e.FilterHash, e.Filter)
}
