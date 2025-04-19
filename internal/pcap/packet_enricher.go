package pcap

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
)

func ConvertPacketToEvent(pkt gopacket.Packet) *pbAgent.PacketEvent {
	metadata := pkt.Metadata()
	event := &pbAgent.PacketEvent{
		CaptureLength:  uint32(metadata.CaptureLength),
		OriginalLength: uint32(metadata.Length),
		InterfaceIndex: int32(metadata.InterfaceIndex),
		Truncated:      metadata.Truncated,
		Layers:         &pbAgent.Layers{},
	}

	// IP layer
	if ipv4Layer := pkt.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ipv4 := ipv4Layer.(*layers.IPv4)
		event.Layers.IpLayer = &pbAgent.IPLayer{
			Version:  "IPv4",
			SrcIp:    ipv4.SrcIP.String(),
			DstIp:    ipv4.DstIP.String(),
			Ttl:      uint32(ipv4.TTL),
			Protocol: uint32(ipv4.Protocol),
		}
	} else if ipv6Layer := pkt.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
		ipv6 := ipv6Layer.(*layers.IPv6)
		event.Layers.IpLayer = &pbAgent.IPLayer{
			Version:  "IPv6",
			SrcIp:    ipv6.SrcIP.String(),
			DstIp:    ipv6.DstIP.String(),
			HopLimit: uint32(ipv6.HopLimit),
			Protocol: uint32(ipv6.NextHeader),
		}
	}

	// TCP layer
	if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		event.Layers.TcpLayer = &pbAgent.TCPLayer{
			SrcPort: uint32(tcp.SrcPort),
			DstPort: uint32(tcp.DstPort),
			Seq:     tcp.Seq,
			Ack:     tcp.Ack,
			Fin:     tcp.FIN,
			Syn:     tcp.SYN,
			Rst:     tcp.RST,
			Psh:     tcp.PSH,
			AckFlag: tcp.ACK,
			Urg:     tcp.URG,
			Window:  uint32(tcp.Window),
		}
	}

	// UDP layer
	if udpLayer := pkt.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		event.Layers.UdpLayer = &pbAgent.UDPLayer{
			SrcPort: uint32(udp.SrcPort),
			DstPort: uint32(udp.DstPort),
			Length:  uint32(udp.Length),
		}
	}

	// TLS layer
	if tlsLayer := pkt.Layer(layers.LayerTypeTLS); tlsLayer != nil {
		tls := tlsLayer.(*layers.TLS)
		tlsInfo := &pbAgent.TLSLayer{}

		for _, hs := range tls.Handshake {
			tlsInfo.Records = append(tlsInfo.Records, &pbAgent.TLSRecord{
				Type:    "Handshake",
				Version: hs.Version.String(),
				Length:  uint32(hs.Length),
			})
		}
		for _, app := range tls.AppData {
			tlsInfo.Records = append(tlsInfo.Records, &pbAgent.TLSRecord{
				Type:    "AppData",
				Version: app.Version.String(),
				Length:  uint32(app.Length),
			})
		}
		for _, alert := range tls.Alert {
			tlsInfo.Records = append(tlsInfo.Records, &pbAgent.TLSRecord{
				Type:    "Alert",
				Version: alert.Version.String(),
				Length:  uint32(alert.Length),
			})
		}
		for _, cc := range tls.ChangeCipherSpec {
			tlsInfo.Records = append(tlsInfo.Records, &pbAgent.TLSRecord{
				Type:    "ChangeCipherSpec",
				Version: cc.Version.String(),
				Length:  uint32(cc.Length),
			})
		}

		event.Layers.TlsLayer = tlsInfo
	}

	return event
}
