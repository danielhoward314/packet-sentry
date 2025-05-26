package dao

type Event struct {
	EventTime      string `json:"event_time,omitempty"`
	Bpf            string `json:"bpf,omitempty"`
	OriginalLength int32  `json:"original_length,omitempty"`
	IpSrc          string `json:"ip_src,omitempty"`
	IpDst          string `json:"ip_dst,omitempty"`
	TcpSrcPort     string `json:"tcp_src_port,omitempty"`
	TcpDstPort     string `json:"tcp_dst_port,omitempty"`
	IpVersion      string `json:"ip_version,omitempty"`
}

type Events interface {
	Read(deviceID string, start string, end string) ([]*Event, error)
}
