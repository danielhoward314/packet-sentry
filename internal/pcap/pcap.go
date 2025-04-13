package pcap

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

// packetCapture holds the config used to create a capture and the handle to the live capture
type packetCapture struct {
	cancelFunc context.CancelFunc
	config     *CaptureConfig
	ctx        context.Context
	handle     *pcap.Handle
	logger     *slog.Logger
	packetOut  chan<- gopacket.Packet
	wg         *sync.WaitGroup
}

func newPacketCapture(parentCtx context.Context, parentLogger *slog.Logger, config *CaptureConfig, wg *sync.WaitGroup, packetOut chan<- gopacket.Packet) (*packetCapture, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	childLogger := parentLogger.With(psLog.KeyCaptureConfig, config)

	return &packetCapture{
		config:     config,
		ctx:        ctx,
		cancelFunc: cancel,
		logger:     childLogger,
		packetOut:  packetOut,
		wg:         wg,
	}, nil
}

// Start begins the packet capture process
func (pc *packetCapture) Start() error {
	logger := pc.logger.With(psLog.KeyFunction, "packetCapture.Start")
	var err error
	pc.wg.Add(1)

	logger.Info("opening live packet capture")
	pc.handle, err = pcap.OpenLive(
		pc.config.DeviceName,
		pc.config.SnapLen,
		pc.config.Promiscuous,
		pc.config.Timeout,
	)
	if err != nil {
		logger.Error("error opening device", psLog.KeyError, err)
		pc.cleanup()
		return err
	}

	logger.Info("setting BPF")
	err = pc.handle.SetBPFFilter(pc.config.BPF)
	if err != nil {
		logger.Error("error setting BPF", psLog.KeyError, err)
		pc.cleanup()
		return err
	}

	go func() {
		// will be called when context is canceled and we exit this goroutine
		defer pc.cleanup()
		packetSource := gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
		packetChan := packetSource.Packets()
		logger.Info("packet source created, entering capture loop")

		for {
			select {
			// ok (true): channel is still open and the read succeeded
			// ok (false): channel has been closed and there's no more data to read
			case packet, ok := <-packetChan:
				if !ok {
					logger.Info("packet channel closed")
					return
				}

				// nested select for handling buffered channel backpressure
				// send does not block on a buffered channel until it's full
				// a nested select here falls into the default case when the channel is full
				// so we can keep processing
				select {
				case pc.packetOut <- packet:
				default:
					// TODO: replace with a dropped packet log that gets sent, then wiped, on a 24-hour interval
					logger.Warn("packet channel full, dropping packet", psLog.KeyDroppedPacket, packet.String())
				}

			case <-pc.ctx.Done():
				logger.Info("context canceled, stopping packet capture")
				return
			}
		}
	}()

	logger.Info("Packet capture started successfully")
	return nil
}

// Stop terminates the packet capture process
func (pc *packetCapture) Stop() {
	logger := pc.logger.With(psLog.KeyFunction, "packetCapture.Stop")
	logger.Info("calling packet capture's cancel func")
	pc.cancelFunc()
}

func (pc *packetCapture) cleanup() {
	logger := pc.logger.With(psLog.KeyFunction, "packetCapture.cleanup")
	if pc.handle != nil {
		logger.Info("closing packet capture handle")
		pc.handle.Close()
		pc.handle = nil
	} else {
		logger.Info("cleanup did not find handle for packet capture")
	}
	pc.wg.Done()
}
