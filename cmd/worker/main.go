package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
)

var protocolNames = []string{
	"HOPOPT",             // 0
	"ICMP",               // 1
	"IGMP",               // 2
	"GGP",                // 3
	"IPv4",               // 4
	"ST",                 // 5
	"TCP",                // 6
	"CBT",                // 7
	"EGP",                // 8
	"IGP",                // 9
	"BBN-RCC-MON",        // 10
	"NVP-II",             // 11
	"PUP",                // 12
	"ARGUS (deprecated)", // 13
	"EMCON",              // 14
	"XNET",               // 15
	"CHAOS",              // 16
	"UDP",                // 17
	"MUX",                // 18
	"DCN-MEAS",           // 19
	"HMP",                // 20
	"PRM",                // 21
	"XNS-IDP",            // 22
	"TRUNK-1",            // 23
	"TRUNK-2",            // 24
	"LEAF-1",             // 25
	"LEAF-2",             // 26
	"RDP",                // 27
	"IRTP",               // 28
	"ISO-TP4",            // 29
	"NETBLT",             // 30
	"MFE-NSP",            // 31
	"MERIT-INP",          // 32
	"DCCP",               // 33
	"3PC",                // 34
	"IDPR",               // 35
	"XTP",                // 36
	"DDP",                // 37
	"IDPR-CMTP",          // 38
	"TP++",               // 39
	"IL",                 // 40
	"IPv6",               // 41
	"SDRP",               // 42
	"IPv6-Route",         // 43
	"IPv6-Frag",          // 44
	"IDRP",               // 45
	"RSVP",               // 46
	"GRE",                // 47
	"DSR",                // 48
	"BNA",                // 49
	"ESP",                // 50
	"AH",                 // 51
	"I-NLSP",             // 52
	"SWIPE (deprecated)", // 53
	"NARP",               // 54
	"Mobile",             // 55
	"TLSP",               // 56
	"SKIP",               // 57
	"IPv6-ICMP",          // 58
	"IPv6-NoNxt",         // 59
	"IPv6-Opts",          // 60
	"",                   // 61 (Unassigned)
	"CFTP",               // 62
	"",                   // 63 (Unassigned)
	"SAT-EXPAK",          // 64
	"KRYPTOLAN",          // 65
	"RVD",                // 66
	"IPPC",               // 67
	"",                   // 68 (Unassigned)
	"SAT-MON",            // 69
	"VISA",               // 70
	"IPCV",               // 71
	"CPNX",               // 72
	"CPHB",               // 73
	"WSN",                // 74
	"PVP",                // 75
	"BR-SAT-MON",         // 76
	"SUN-ND",             // 77
	"WB-MON",             // 78
	"WB-EXPAK",           // 79
	"ISO-IP",             // 80
	"VMTP",               // 81
	"SECURE-VMTP",        // 82
	"VINES",              // 83
	"TTP",                // 84
	"NSFNET-IGP",         // 85
	"DGP",                // 86
	"TCF",                // 87
	"EIGRP",              // 88
	"OSPFIGP",            // 89
	"Sprite-RPC",         // 90
	"LARP",               // 91
	"MTP",                // 92
	"AX.25",              // 93
	"IPIP",               // 94
	"MICP (deprecated)",  // 95
	"SCC-SP",             // 96
	"ETHERIP",            // 97
	"ENCAP",              // 98
	"",                   // 99 (Private Use)
	"GMB",                // 100
	"IFMP",               // 101
	"PNNI",               // 102
	"PIM",                // 103
	"ARIS",               // 104
	"SCPS",               // 105
	"QNX",                // 106
	"A/N",                // 107
	"IPComp",             // 108
	"SNP",                // 109
	"Compaq-Peer",        // 110
	"IPX-in-IP",          // 111
	"VRRP",               // 112
	"PGM",                // 113
	"",                   // 114 (Unassigned)
	"L2TP",               // 115
	"DDX",                // 116
	"IATP",               // 117
	"STP",                // 118
	"SRP",                // 119
	"UTI",                // 120
	"SMP",                // 121
	"SM (deprecated)",    // 122
	"PTP",                // 123
	"ISIS over IPv4",     // 124
	"FIRE",               // 125
	"CRTP",               // 126
	"CRUDP",              // 127
	"SSCOPMCE",           // 128
	"IPLT",               // 129
	"SPS",                // 130
	"PIPE",               // 131
	"SCTP",               // 132
	"FC",                 // 133
	"RSVP-E2E-IGNORE",    // 134
	"Mobility Header",    // 135
	"UDPLite",            // 136
	"MPLS-in-IP",         // 137
	"manet",              // 138
	"HIP",                // 139
	"Shim6",              // 140
	"WESP",               // 141
	"ROHC",               // 142
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	dbHost := getEnv("TSDB_HOST", "localhost")
	dbPort := getEnv("TSDB_PORT", "5432")
	dbUser := getEnv("TSDB_USER", "postgres")
	dbPassword := getEnv("TSDB_PASSWORD", "")
	dbName := getEnv("TSDB_DATABASE", "packets")
	dbSSLMode := getEnv("TSDB_SSLMODE", "disable")
	dbConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal("Error connecting to TimescaleDB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("TimescaleDB ping failed:", err)
	}

	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events.*"},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse && !strings.Contains(err.Error(), "already in use") {
		log.Fatal("AddStream error:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		logger.Info("Received shutdown signal, exiting...")
		cancel()
	}()

	sub, err := js.Subscribe("events.*", func(msg *nats.Msg) {
		handlePacketEvent(ctx, logger, db, msg)
	}, nats.Durable("worker-durable"), nats.ManualAck())
	if err != nil {
		log.Fatal("Error subscribing to JetStream:", err)
	}
	defer sub.Unsubscribe()

	logger.Info("Worker started, listening for packet events...")

	<-ctx.Done()
}

func handlePacketEvent(ctx context.Context, logger *slog.Logger, db *sql.DB, msg *nats.Msg) {
	var packetEvent pbAgent.PacketEvent
	err := proto.Unmarshal(msg.Data, &packetEvent)
	if err != nil {
		logger.Info("Failed to unmarshal protobuf message:", "error", err)
		_ = msg.Ack()
		return
	}

	// the NATS subject should be "events.id" where the id is the device row's os_unique_identifier
	osUniqueIdentifier := ""
	parts := strings.SplitN(msg.Subject, ".", 2)
	if len(parts) == 2 {
		osUniqueIdentifier = parts[1]
	}
	logger.Info("derived os_unique_identifier from subject", "os_unique_identifier", osUniqueIdentifier)

	// capture config
	bpf := packetEvent.Bpf
	interfaceName := packetEvent.DeviceName
	promiscuous := packetEvent.Promiscuous
	snapLen := packetEvent.SnapLen

	// capture metadata
	captureLen := int32(packetEvent.CaptureLength)
	originalLen := int32(packetEvent.OriginalLength)
	truncated := packetEvent.Truncated
	interfaceIndex := packetEvent.InterfaceIndex

	// ipLayer
	var dstIP, ipVersion, ipProtocol, srcIP string
	var ipHopLimit, ipTTL int32

	// tcpLayer
	var tcpAckFlag, tcpFin, tcpPsh, tcpRst, tcpSyn, tcpUrg bool
	var dstPortTCP, srcPortTCP, tcpWindow int32
	var tcpAck, tcpSeq int64

	var dstPortUDP, srcPortUDP, udpLen int32
	var tlsRecordsCount int

	if packetEvent.Layers != nil {
		if packetEvent.Layers.IpLayer != nil {
			srcIP = packetEvent.Layers.IpLayer.SrcIp
			dstIP = packetEvent.Layers.IpLayer.DstIp
			ipVersion = packetEvent.Layers.IpLayer.Version
			if ipVersion == "IPv4" {
				ipProtocol = getProtocolName(packetEvent.Layers.IpLayer.Protocol)
				ipTTL = int32(packetEvent.Layers.IpLayer.Ttl)
			} else {
				ipHopLimit = int32(packetEvent.Layers.IpLayer.HopLimit)
			}
		}

		if packetEvent.Layers.TcpLayer != nil {
			srcPortTCP = int32(packetEvent.Layers.TcpLayer.SrcPort)
			dstPortTCP = int32(packetEvent.Layers.TcpLayer.DstPort)
			tcpAck = int64(packetEvent.Layers.TcpLayer.Ack)
			tcpSeq = int64(packetEvent.Layers.TcpLayer.Seq)
			tcpWindow = int32(packetEvent.Layers.TcpLayer.Window)

			tcpAckFlag = packetEvent.Layers.TcpLayer.AckFlag
			tcpFin = packetEvent.Layers.TcpLayer.Fin
			tcpPsh = packetEvent.Layers.TcpLayer.Psh
			tcpRst = packetEvent.Layers.TcpLayer.Rst
			tcpSyn = packetEvent.Layers.TcpLayer.Syn
			tcpUrg = packetEvent.Layers.TcpLayer.Urg
		} else if packetEvent.Layers.UdpLayer != nil {
			srcPortUDP = int32(packetEvent.Layers.UdpLayer.SrcPort)
			dstPortUDP = int32(packetEvent.Layers.UdpLayer.DstPort)
			udpLen = int32(packetEvent.Layers.UdpLayer.Length)
		} else if packetEvent.Layers.TlsLayer != nil {
			tlsRecordsCount = len(packetEvent.Layers.TlsLayer.Records)
		}
	}

	debugSQL := fmt.Sprintf(
		`INSERT INTO packet_events (
            os_unique_identifier, bpf, interface, promiscuous, snap_length,
            capture_length, original_length, interface_index, truncated, ip_version,
            ip_src, ip_dst, ip_ttl, ip_hop_limit, ip_protocol,
            tcp_src_port, tcp_dst_port, tcp_seq, tcp_ack, tcp_fin,
            tcp_syn, tcp_rst, tcp_psh, tcp_ack_flag, tcp_urg,
            tcp_window, udp_src_port, udp_dst_port, udp_length, tls_record_count
        ) VALUES (
            '%v', '%v', '%v', %v, %v,
            %v, %v, %v, %v, '%v',
            '%v', '%v', %v, %v, '%v',
            %v, %v, %v, %v, %v,
            %v, %v, %v, %v, %v,
            %v, %v, %v, %v, %v
        )`,
		osUniqueIdentifier, bpf, interfaceName, promiscuous, snapLen,
		captureLen, originalLen, interfaceIndex, truncated, ipVersion,
		srcIP, dstIP, ipTTL, ipHopLimit, ipProtocol,
		srcPortTCP, dstPortTCP, tcpSeq, tcpAck, tcpFin,
		tcpSyn, tcpRst, tcpPsh, tcpAckFlag, tcpUrg,
		tcpWindow, srcPortUDP, dstPortUDP, udpLen, int32(tlsRecordsCount),
	)
	logger.Info("Debug SQL query", "sql", debugSQL)

	query := `
	INSERT INTO packet_events (
		os_unique_identifier, bpf, interface, promiscuous, snap_length,
		capture_length, original_length, interface_index, truncated, ip_version,
		ip_src, ip_dst, ip_ttl, ip_hop_limit, ip_protocol,
		tcp_src_port, tcp_dst_port, tcp_seq, tcp_ack, tcp_fin,
		tcp_syn, tcp_rst, tcp_psh, tcp_ack_flag, tcp_urg,
		tcp_window, udp_src_port, udp_dst_port, udp_length, tls_record_count
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15,
		$16, $17, $18, $19, $20,
		$21, $22, $23, $24, $25,
		$26, $27, $28, $29, $30
	)
	RETURNING id, event_time;
	`

	var id int
	var eventTime time.Time

	err = db.QueryRowContext(
		ctx,
		query,
		osUniqueIdentifier, bpf, interfaceName, promiscuous, snapLen, // $1 - $5
		captureLen, originalLen, interfaceIndex, truncated, ipVersion, // $6 - $10
		srcIP, dstIP, ipTTL, ipHopLimit, ipProtocol, // $11 - $15
		srcPortTCP, dstPortTCP, tcpSeq, tcpAck, tcpFin, // $16 - $20
		tcpSyn, tcpRst, tcpPsh, tcpAckFlag, tcpUrg, // $21 - $25
		tcpWindow, srcPortUDP, dstPortUDP, udpLen, int32(tlsRecordsCount), // $26 - $30
	).Scan(&id, &eventTime)
	if err != nil {
		log.Printf("insert error: %v", err)
	}

	_ = msg.Ack()
}

// getEnv reads an environment variable or returns a default
func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getProtocolName(proto uint32) string {
	if proto < uint32(len(protocolNames)) && protocolNames[proto] != "" {
		return protocolNames[proto]
	}
	return "Unknown Protocol"
}
