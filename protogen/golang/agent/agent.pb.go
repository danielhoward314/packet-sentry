// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.12.4
// source: agent/agent.proto

package agent

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Empty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_agent_agent_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{0}
}

type InterfaceDetails struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InterfaceDetails) Reset() {
	*x = InterfaceDetails{}
	mi := &file_agent_agent_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InterfaceDetails) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InterfaceDetails) ProtoMessage() {}

func (x *InterfaceDetails) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InterfaceDetails.ProtoReflect.Descriptor instead.
func (*InterfaceDetails) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{1}
}

func (x *InterfaceDetails) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ReportInterfacesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Interfaces    []*InterfaceDetails    `protobuf:"bytes,1,rep,name=interfaces,proto3" json:"interfaces,omitempty"`
	PcapVersion   string                 `protobuf:"bytes,2,opt,name=pcapVersion,proto3" json:"pcapVersion,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReportInterfacesRequest) Reset() {
	*x = ReportInterfacesRequest{}
	mi := &file_agent_agent_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReportInterfacesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportInterfacesRequest) ProtoMessage() {}

func (x *ReportInterfacesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportInterfacesRequest.ProtoReflect.Descriptor instead.
func (*ReportInterfacesRequest) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{2}
}

func (x *ReportInterfacesRequest) GetInterfaces() []*InterfaceDetails {
	if x != nil {
		return x.Interfaces
	}
	return nil
}

func (x *ReportInterfacesRequest) GetPcapVersion() string {
	if x != nil {
		return x.PcapVersion
	}
	return ""
}

type Command struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Command) Reset() {
	*x = Command{}
	mi := &file_agent_agent_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{3}
}

func (x *Command) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type CommandsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Commands      []*Command             `protobuf:"bytes,1,rep,name=commands,proto3" json:"commands,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CommandsResponse) Reset() {
	*x = CommandsResponse{}
	mi := &file_agent_agent_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CommandsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommandsResponse) ProtoMessage() {}

func (x *CommandsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommandsResponse.ProtoReflect.Descriptor instead.
func (*CommandsResponse) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{4}
}

func (x *CommandsResponse) GetCommands() []*Command {
	if x != nil {
		return x.Commands
	}
	return nil
}

type CaptureConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Bpf           string                 `protobuf:"bytes,1,opt,name=bpf,proto3" json:"bpf,omitempty"`
	DeviceName    string                 `protobuf:"bytes,2,opt,name=deviceName,proto3" json:"deviceName,omitempty"`
	Promiscuous   bool                   `protobuf:"varint,3,opt,name=promiscuous,proto3" json:"promiscuous,omitempty"`
	SnapLen       int32                  `protobuf:"varint,4,opt,name=snapLen,proto3" json:"snapLen,omitempty"`
	Timeout       int64                  `protobuf:"varint,5,opt,name=timeout,proto3" json:"timeout,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CaptureConfig) Reset() {
	*x = CaptureConfig{}
	mi := &file_agent_agent_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CaptureConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CaptureConfig) ProtoMessage() {}

func (x *CaptureConfig) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CaptureConfig.ProtoReflect.Descriptor instead.
func (*CaptureConfig) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{5}
}

func (x *CaptureConfig) GetBpf() string {
	if x != nil {
		return x.Bpf
	}
	return ""
}

func (x *CaptureConfig) GetDeviceName() string {
	if x != nil {
		return x.DeviceName
	}
	return ""
}

func (x *CaptureConfig) GetPromiscuous() bool {
	if x != nil {
		return x.Promiscuous
	}
	return false
}

func (x *CaptureConfig) GetSnapLen() int32 {
	if x != nil {
		return x.SnapLen
	}
	return 0
}

func (x *CaptureConfig) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

type BPFConfig struct {
	state         protoimpl.MessageState          `protogen:"open.v1"`
	Create        map[string]*InterfaceCaptureMap `protobuf:"bytes,1,rep,name=create,proto3" json:"create,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Update        map[string]*InterfaceCaptureMap `protobuf:"bytes,2,rep,name=update,proto3" json:"update,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Delete        map[string]*InterfaceCaptureMap `protobuf:"bytes,3,rep,name=delete,proto3" json:"delete,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BPFConfig) Reset() {
	*x = BPFConfig{}
	mi := &file_agent_agent_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BPFConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BPFConfig) ProtoMessage() {}

func (x *BPFConfig) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BPFConfig.ProtoReflect.Descriptor instead.
func (*BPFConfig) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{6}
}

func (x *BPFConfig) GetCreate() map[string]*InterfaceCaptureMap {
	if x != nil {
		return x.Create
	}
	return nil
}

func (x *BPFConfig) GetUpdate() map[string]*InterfaceCaptureMap {
	if x != nil {
		return x.Update
	}
	return nil
}

func (x *BPFConfig) GetDelete() map[string]*InterfaceCaptureMap {
	if x != nil {
		return x.Delete
	}
	return nil
}

type InterfaceCaptureMap struct {
	state         protoimpl.MessageState    `protogen:"open.v1"`
	Captures      map[uint64]*CaptureConfig `protobuf:"bytes,1,rep,name=captures,proto3" json:"captures,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InterfaceCaptureMap) Reset() {
	*x = InterfaceCaptureMap{}
	mi := &file_agent_agent_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InterfaceCaptureMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InterfaceCaptureMap) ProtoMessage() {}

func (x *InterfaceCaptureMap) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InterfaceCaptureMap.ProtoReflect.Descriptor instead.
func (*InterfaceCaptureMap) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{7}
}

func (x *InterfaceCaptureMap) GetCaptures() map[uint64]*CaptureConfig {
	if x != nil {
		return x.Captures
	}
	return nil
}

type PacketEvent struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Bpf            string                 `protobuf:"bytes,1,opt,name=bpf,proto3" json:"bpf,omitempty"`
	DeviceName     string                 `protobuf:"bytes,2,opt,name=deviceName,proto3" json:"deviceName,omitempty"`
	Promiscuous    bool                   `protobuf:"varint,3,opt,name=promiscuous,proto3" json:"promiscuous,omitempty"`
	SnapLen        int32                  `protobuf:"varint,4,opt,name=snapLen,proto3" json:"snapLen,omitempty"`
	CaptureLength  uint32                 `protobuf:"varint,5,opt,name=capture_length,json=captureLength,proto3" json:"capture_length,omitempty"`
	OriginalLength uint32                 `protobuf:"varint,6,opt,name=original_length,json=originalLength,proto3" json:"original_length,omitempty"`
	InterfaceIndex int32                  `protobuf:"varint,7,opt,name=interface_index,json=interfaceIndex,proto3" json:"interface_index,omitempty"`
	Truncated      bool                   `protobuf:"varint,8,opt,name=truncated,proto3" json:"truncated,omitempty"`
	Layers         *Layers                `protobuf:"bytes,9,opt,name=layers,proto3" json:"layers,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *PacketEvent) Reset() {
	*x = PacketEvent{}
	mi := &file_agent_agent_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PacketEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PacketEvent) ProtoMessage() {}

func (x *PacketEvent) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PacketEvent.ProtoReflect.Descriptor instead.
func (*PacketEvent) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{8}
}

func (x *PacketEvent) GetBpf() string {
	if x != nil {
		return x.Bpf
	}
	return ""
}

func (x *PacketEvent) GetDeviceName() string {
	if x != nil {
		return x.DeviceName
	}
	return ""
}

func (x *PacketEvent) GetPromiscuous() bool {
	if x != nil {
		return x.Promiscuous
	}
	return false
}

func (x *PacketEvent) GetSnapLen() int32 {
	if x != nil {
		return x.SnapLen
	}
	return 0
}

func (x *PacketEvent) GetCaptureLength() uint32 {
	if x != nil {
		return x.CaptureLength
	}
	return 0
}

func (x *PacketEvent) GetOriginalLength() uint32 {
	if x != nil {
		return x.OriginalLength
	}
	return 0
}

func (x *PacketEvent) GetInterfaceIndex() int32 {
	if x != nil {
		return x.InterfaceIndex
	}
	return 0
}

func (x *PacketEvent) GetTruncated() bool {
	if x != nil {
		return x.Truncated
	}
	return false
}

func (x *PacketEvent) GetLayers() *Layers {
	if x != nil {
		return x.Layers
	}
	return nil
}

type Layers struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	IpLayer       *IPLayer               `protobuf:"bytes,1,opt,name=ip_layer,json=ipLayer,proto3" json:"ip_layer,omitempty"`
	TcpLayer      *TCPLayer              `protobuf:"bytes,2,opt,name=tcp_layer,json=tcpLayer,proto3" json:"tcp_layer,omitempty"`
	UdpLayer      *UDPLayer              `protobuf:"bytes,3,opt,name=udp_layer,json=udpLayer,proto3" json:"udp_layer,omitempty"`
	TlsLayer      *TLSLayer              `protobuf:"bytes,4,opt,name=tls_layer,json=tlsLayer,proto3" json:"tls_layer,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Layers) Reset() {
	*x = Layers{}
	mi := &file_agent_agent_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Layers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Layers) ProtoMessage() {}

func (x *Layers) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Layers.ProtoReflect.Descriptor instead.
func (*Layers) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{9}
}

func (x *Layers) GetIpLayer() *IPLayer {
	if x != nil {
		return x.IpLayer
	}
	return nil
}

func (x *Layers) GetTcpLayer() *TCPLayer {
	if x != nil {
		return x.TcpLayer
	}
	return nil
}

func (x *Layers) GetUdpLayer() *UDPLayer {
	if x != nil {
		return x.UdpLayer
	}
	return nil
}

func (x *Layers) GetTlsLayer() *TLSLayer {
	if x != nil {
		return x.TlsLayer
	}
	return nil
}

type IPLayer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Version       string                 `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"` // "IPv4" or "IPv6"
	SrcIp         string                 `protobuf:"bytes,2,opt,name=src_ip,json=srcIp,proto3" json:"src_ip,omitempty"`
	DstIp         string                 `protobuf:"bytes,3,opt,name=dst_ip,json=dstIp,proto3" json:"dst_ip,omitempty"`
	Ttl           uint32                 `protobuf:"varint,4,opt,name=ttl,proto3" json:"ttl,omitempty"`                           // For IPv4
	HopLimit      uint32                 `protobuf:"varint,5,opt,name=hop_limit,json=hopLimit,proto3" json:"hop_limit,omitempty"` // For IPv6
	Protocol      uint32                 `protobuf:"varint,6,opt,name=protocol,proto3" json:"protocol,omitempty"`                 // L4 protocol number
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IPLayer) Reset() {
	*x = IPLayer{}
	mi := &file_agent_agent_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IPLayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPLayer) ProtoMessage() {}

func (x *IPLayer) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPLayer.ProtoReflect.Descriptor instead.
func (*IPLayer) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{10}
}

func (x *IPLayer) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *IPLayer) GetSrcIp() string {
	if x != nil {
		return x.SrcIp
	}
	return ""
}

func (x *IPLayer) GetDstIp() string {
	if x != nil {
		return x.DstIp
	}
	return ""
}

func (x *IPLayer) GetTtl() uint32 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

func (x *IPLayer) GetHopLimit() uint32 {
	if x != nil {
		return x.HopLimit
	}
	return 0
}

func (x *IPLayer) GetProtocol() uint32 {
	if x != nil {
		return x.Protocol
	}
	return 0
}

type TCPLayer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SrcPort       uint32                 `protobuf:"varint,1,opt,name=src_port,json=srcPort,proto3" json:"src_port,omitempty"`
	DstPort       uint32                 `protobuf:"varint,2,opt,name=dst_port,json=dstPort,proto3" json:"dst_port,omitempty"`
	Seq           uint32                 `protobuf:"varint,3,opt,name=seq,proto3" json:"seq,omitempty"`
	Ack           uint32                 `protobuf:"varint,4,opt,name=ack,proto3" json:"ack,omitempty"`
	Fin           bool                   `protobuf:"varint,5,opt,name=fin,proto3" json:"fin,omitempty"`
	Syn           bool                   `protobuf:"varint,6,opt,name=syn,proto3" json:"syn,omitempty"`
	Rst           bool                   `protobuf:"varint,7,opt,name=rst,proto3" json:"rst,omitempty"`
	Psh           bool                   `protobuf:"varint,8,opt,name=psh,proto3" json:"psh,omitempty"`
	AckFlag       bool                   `protobuf:"varint,9,opt,name=ack_flag,json=ackFlag,proto3" json:"ack_flag,omitempty"`
	Urg           bool                   `protobuf:"varint,10,opt,name=urg,proto3" json:"urg,omitempty"`
	Window        uint32                 `protobuf:"varint,11,opt,name=window,proto3" json:"window,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TCPLayer) Reset() {
	*x = TCPLayer{}
	mi := &file_agent_agent_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TCPLayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TCPLayer) ProtoMessage() {}

func (x *TCPLayer) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TCPLayer.ProtoReflect.Descriptor instead.
func (*TCPLayer) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{11}
}

func (x *TCPLayer) GetSrcPort() uint32 {
	if x != nil {
		return x.SrcPort
	}
	return 0
}

func (x *TCPLayer) GetDstPort() uint32 {
	if x != nil {
		return x.DstPort
	}
	return 0
}

func (x *TCPLayer) GetSeq() uint32 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *TCPLayer) GetAck() uint32 {
	if x != nil {
		return x.Ack
	}
	return 0
}

func (x *TCPLayer) GetFin() bool {
	if x != nil {
		return x.Fin
	}
	return false
}

func (x *TCPLayer) GetSyn() bool {
	if x != nil {
		return x.Syn
	}
	return false
}

func (x *TCPLayer) GetRst() bool {
	if x != nil {
		return x.Rst
	}
	return false
}

func (x *TCPLayer) GetPsh() bool {
	if x != nil {
		return x.Psh
	}
	return false
}

func (x *TCPLayer) GetAckFlag() bool {
	if x != nil {
		return x.AckFlag
	}
	return false
}

func (x *TCPLayer) GetUrg() bool {
	if x != nil {
		return x.Urg
	}
	return false
}

func (x *TCPLayer) GetWindow() uint32 {
	if x != nil {
		return x.Window
	}
	return 0
}

type UDPLayer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SrcPort       uint32                 `protobuf:"varint,1,opt,name=src_port,json=srcPort,proto3" json:"src_port,omitempty"`
	DstPort       uint32                 `protobuf:"varint,2,opt,name=dst_port,json=dstPort,proto3" json:"dst_port,omitempty"`
	Length        uint32                 `protobuf:"varint,3,opt,name=length,proto3" json:"length,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UDPLayer) Reset() {
	*x = UDPLayer{}
	mi := &file_agent_agent_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UDPLayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UDPLayer) ProtoMessage() {}

func (x *UDPLayer) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UDPLayer.ProtoReflect.Descriptor instead.
func (*UDPLayer) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{12}
}

func (x *UDPLayer) GetSrcPort() uint32 {
	if x != nil {
		return x.SrcPort
	}
	return 0
}

func (x *UDPLayer) GetDstPort() uint32 {
	if x != nil {
		return x.DstPort
	}
	return 0
}

func (x *UDPLayer) GetLength() uint32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type TLSLayer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Records       []*TLSRecord           `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TLSLayer) Reset() {
	*x = TLSLayer{}
	mi := &file_agent_agent_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TLSLayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TLSLayer) ProtoMessage() {}

func (x *TLSLayer) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TLSLayer.ProtoReflect.Descriptor instead.
func (*TLSLayer) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{13}
}

func (x *TLSLayer) GetRecords() []*TLSRecord {
	if x != nil {
		return x.Records
	}
	return nil
}

type TLSRecord struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          string                 `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Length        uint32                 `protobuf:"varint,3,opt,name=length,proto3" json:"length,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TLSRecord) Reset() {
	*x = TLSRecord{}
	mi := &file_agent_agent_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TLSRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TLSRecord) ProtoMessage() {}

func (x *TLSRecord) ProtoReflect() protoreflect.Message {
	mi := &file_agent_agent_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TLSRecord.ProtoReflect.Descriptor instead.
func (*TLSRecord) Descriptor() ([]byte, []int) {
	return file_agent_agent_proto_rawDescGZIP(), []int{14}
}

func (x *TLSRecord) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *TLSRecord) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *TLSRecord) GetLength() uint32 {
	if x != nil {
		return x.Length
	}
	return 0
}

var File_agent_agent_proto protoreflect.FileDescriptor

const file_agent_agent_proto_rawDesc = "" +
	"\n" +
	"\x11agent/agent.proto\x12\x05agent\"\a\n" +
	"\x05Empty\"&\n" +
	"\x10InterfaceDetails\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"t\n" +
	"\x17ReportInterfacesRequest\x127\n" +
	"\n" +
	"interfaces\x18\x01 \x03(\v2\x17.agent.InterfaceDetailsR\n" +
	"interfaces\x12 \n" +
	"\vpcapVersion\x18\x02 \x01(\tR\vpcapVersion\"\x1d\n" +
	"\aCommand\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\">\n" +
	"\x10CommandsResponse\x12*\n" +
	"\bcommands\x18\x01 \x03(\v2\x0e.agent.CommandR\bcommands\"\x97\x01\n" +
	"\rCaptureConfig\x12\x10\n" +
	"\x03bpf\x18\x01 \x01(\tR\x03bpf\x12\x1e\n" +
	"\n" +
	"deviceName\x18\x02 \x01(\tR\n" +
	"deviceName\x12 \n" +
	"\vpromiscuous\x18\x03 \x01(\bR\vpromiscuous\x12\x18\n" +
	"\asnapLen\x18\x04 \x01(\x05R\asnapLen\x12\x18\n" +
	"\atimeout\x18\x05 \x01(\x03R\atimeout\"\xb2\x03\n" +
	"\tBPFConfig\x124\n" +
	"\x06create\x18\x01 \x03(\v2\x1c.agent.BPFConfig.CreateEntryR\x06create\x124\n" +
	"\x06update\x18\x02 \x03(\v2\x1c.agent.BPFConfig.UpdateEntryR\x06update\x124\n" +
	"\x06delete\x18\x03 \x03(\v2\x1c.agent.BPFConfig.DeleteEntryR\x06delete\x1aU\n" +
	"\vCreateEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x120\n" +
	"\x05value\x18\x02 \x01(\v2\x1a.agent.InterfaceCaptureMapR\x05value:\x028\x01\x1aU\n" +
	"\vUpdateEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x120\n" +
	"\x05value\x18\x02 \x01(\v2\x1a.agent.InterfaceCaptureMapR\x05value:\x028\x01\x1aU\n" +
	"\vDeleteEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x120\n" +
	"\x05value\x18\x02 \x01(\v2\x1a.agent.InterfaceCaptureMapR\x05value:\x028\x01\"\xae\x01\n" +
	"\x13InterfaceCaptureMap\x12D\n" +
	"\bcaptures\x18\x01 \x03(\v2(.agent.InterfaceCaptureMap.CapturesEntryR\bcaptures\x1aQ\n" +
	"\rCapturesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\x04R\x03key\x12*\n" +
	"\x05value\x18\x02 \x01(\v2\x14.agent.CaptureConfigR\x05value:\x028\x01\"\xb9\x02\n" +
	"\vPacketEvent\x12\x10\n" +
	"\x03bpf\x18\x01 \x01(\tR\x03bpf\x12\x1e\n" +
	"\n" +
	"deviceName\x18\x02 \x01(\tR\n" +
	"deviceName\x12 \n" +
	"\vpromiscuous\x18\x03 \x01(\bR\vpromiscuous\x12\x18\n" +
	"\asnapLen\x18\x04 \x01(\x05R\asnapLen\x12%\n" +
	"\x0ecapture_length\x18\x05 \x01(\rR\rcaptureLength\x12'\n" +
	"\x0foriginal_length\x18\x06 \x01(\rR\x0eoriginalLength\x12'\n" +
	"\x0finterface_index\x18\a \x01(\x05R\x0einterfaceIndex\x12\x1c\n" +
	"\ttruncated\x18\b \x01(\bR\ttruncated\x12%\n" +
	"\x06layers\x18\t \x01(\v2\r.agent.LayersR\x06layers\"\xbd\x01\n" +
	"\x06Layers\x12)\n" +
	"\bip_layer\x18\x01 \x01(\v2\x0e.agent.IPLayerR\aipLayer\x12,\n" +
	"\ttcp_layer\x18\x02 \x01(\v2\x0f.agent.TCPLayerR\btcpLayer\x12,\n" +
	"\tudp_layer\x18\x03 \x01(\v2\x0f.agent.UDPLayerR\budpLayer\x12,\n" +
	"\ttls_layer\x18\x04 \x01(\v2\x0f.agent.TLSLayerR\btlsLayer\"\x9c\x01\n" +
	"\aIPLayer\x12\x18\n" +
	"\aversion\x18\x01 \x01(\tR\aversion\x12\x15\n" +
	"\x06src_ip\x18\x02 \x01(\tR\x05srcIp\x12\x15\n" +
	"\x06dst_ip\x18\x03 \x01(\tR\x05dstIp\x12\x10\n" +
	"\x03ttl\x18\x04 \x01(\rR\x03ttl\x12\x1b\n" +
	"\thop_limit\x18\x05 \x01(\rR\bhopLimit\x12\x1a\n" +
	"\bprotocol\x18\x06 \x01(\rR\bprotocol\"\xf1\x01\n" +
	"\bTCPLayer\x12\x19\n" +
	"\bsrc_port\x18\x01 \x01(\rR\asrcPort\x12\x19\n" +
	"\bdst_port\x18\x02 \x01(\rR\adstPort\x12\x10\n" +
	"\x03seq\x18\x03 \x01(\rR\x03seq\x12\x10\n" +
	"\x03ack\x18\x04 \x01(\rR\x03ack\x12\x10\n" +
	"\x03fin\x18\x05 \x01(\bR\x03fin\x12\x10\n" +
	"\x03syn\x18\x06 \x01(\bR\x03syn\x12\x10\n" +
	"\x03rst\x18\a \x01(\bR\x03rst\x12\x10\n" +
	"\x03psh\x18\b \x01(\bR\x03psh\x12\x19\n" +
	"\back_flag\x18\t \x01(\bR\aackFlag\x12\x10\n" +
	"\x03urg\x18\n" +
	" \x01(\bR\x03urg\x12\x16\n" +
	"\x06window\x18\v \x01(\rR\x06window\"X\n" +
	"\bUDPLayer\x12\x19\n" +
	"\bsrc_port\x18\x01 \x01(\rR\asrcPort\x12\x19\n" +
	"\bdst_port\x18\x02 \x01(\rR\adstPort\x12\x16\n" +
	"\x06length\x18\x03 \x01(\rR\x06length\"6\n" +
	"\bTLSLayer\x12*\n" +
	"\arecords\x18\x01 \x03(\v2\x10.agent.TLSRecordR\arecords\"Q\n" +
	"\tTLSRecord\x12\x12\n" +
	"\x04type\x18\x01 \x01(\tR\x04type\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12\x16\n" +
	"\x06length\x18\x03 \x01(\rR\x06length2\xed\x01\n" +
	"\fAgentService\x12@\n" +
	"\x10ReportInterfaces\x12\x1e.agent.ReportInterfacesRequest\x1a\f.agent.Empty\x125\n" +
	"\x0fSendPacketEvent\x12\x12.agent.PacketEvent\x1a\f.agent.Empty(\x01\x124\n" +
	"\vPollCommand\x12\f.agent.Empty\x1a\x17.agent.CommandsResponse\x12.\n" +
	"\fGetBPFConfig\x12\f.agent.Empty\x1a\x10.agent.BPFConfigB@Z>github.com/danielhoward314/packet-sentry/protogen/golang/agentb\x06proto3"

var (
	file_agent_agent_proto_rawDescOnce sync.Once
	file_agent_agent_proto_rawDescData []byte
)

func file_agent_agent_proto_rawDescGZIP() []byte {
	file_agent_agent_proto_rawDescOnce.Do(func() {
		file_agent_agent_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_agent_agent_proto_rawDesc), len(file_agent_agent_proto_rawDesc)))
	})
	return file_agent_agent_proto_rawDescData
}

var file_agent_agent_proto_msgTypes = make([]protoimpl.MessageInfo, 19)
var file_agent_agent_proto_goTypes = []any{
	(*Empty)(nil),                   // 0: agent.Empty
	(*InterfaceDetails)(nil),        // 1: agent.InterfaceDetails
	(*ReportInterfacesRequest)(nil), // 2: agent.ReportInterfacesRequest
	(*Command)(nil),                 // 3: agent.Command
	(*CommandsResponse)(nil),        // 4: agent.CommandsResponse
	(*CaptureConfig)(nil),           // 5: agent.CaptureConfig
	(*BPFConfig)(nil),               // 6: agent.BPFConfig
	(*InterfaceCaptureMap)(nil),     // 7: agent.InterfaceCaptureMap
	(*PacketEvent)(nil),             // 8: agent.PacketEvent
	(*Layers)(nil),                  // 9: agent.Layers
	(*IPLayer)(nil),                 // 10: agent.IPLayer
	(*TCPLayer)(nil),                // 11: agent.TCPLayer
	(*UDPLayer)(nil),                // 12: agent.UDPLayer
	(*TLSLayer)(nil),                // 13: agent.TLSLayer
	(*TLSRecord)(nil),               // 14: agent.TLSRecord
	nil,                             // 15: agent.BPFConfig.CreateEntry
	nil,                             // 16: agent.BPFConfig.UpdateEntry
	nil,                             // 17: agent.BPFConfig.DeleteEntry
	nil,                             // 18: agent.InterfaceCaptureMap.CapturesEntry
}
var file_agent_agent_proto_depIdxs = []int32{
	1,  // 0: agent.ReportInterfacesRequest.interfaces:type_name -> agent.InterfaceDetails
	3,  // 1: agent.CommandsResponse.commands:type_name -> agent.Command
	15, // 2: agent.BPFConfig.create:type_name -> agent.BPFConfig.CreateEntry
	16, // 3: agent.BPFConfig.update:type_name -> agent.BPFConfig.UpdateEntry
	17, // 4: agent.BPFConfig.delete:type_name -> agent.BPFConfig.DeleteEntry
	18, // 5: agent.InterfaceCaptureMap.captures:type_name -> agent.InterfaceCaptureMap.CapturesEntry
	9,  // 6: agent.PacketEvent.layers:type_name -> agent.Layers
	10, // 7: agent.Layers.ip_layer:type_name -> agent.IPLayer
	11, // 8: agent.Layers.tcp_layer:type_name -> agent.TCPLayer
	12, // 9: agent.Layers.udp_layer:type_name -> agent.UDPLayer
	13, // 10: agent.Layers.tls_layer:type_name -> agent.TLSLayer
	14, // 11: agent.TLSLayer.records:type_name -> agent.TLSRecord
	7,  // 12: agent.BPFConfig.CreateEntry.value:type_name -> agent.InterfaceCaptureMap
	7,  // 13: agent.BPFConfig.UpdateEntry.value:type_name -> agent.InterfaceCaptureMap
	7,  // 14: agent.BPFConfig.DeleteEntry.value:type_name -> agent.InterfaceCaptureMap
	5,  // 15: agent.InterfaceCaptureMap.CapturesEntry.value:type_name -> agent.CaptureConfig
	2,  // 16: agent.AgentService.ReportInterfaces:input_type -> agent.ReportInterfacesRequest
	8,  // 17: agent.AgentService.SendPacketEvent:input_type -> agent.PacketEvent
	0,  // 18: agent.AgentService.PollCommand:input_type -> agent.Empty
	0,  // 19: agent.AgentService.GetBPFConfig:input_type -> agent.Empty
	0,  // 20: agent.AgentService.ReportInterfaces:output_type -> agent.Empty
	0,  // 21: agent.AgentService.SendPacketEvent:output_type -> agent.Empty
	4,  // 22: agent.AgentService.PollCommand:output_type -> agent.CommandsResponse
	6,  // 23: agent.AgentService.GetBPFConfig:output_type -> agent.BPFConfig
	20, // [20:24] is the sub-list for method output_type
	16, // [16:20] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_agent_agent_proto_init() }
func file_agent_agent_proto_init() {
	if File_agent_agent_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_agent_agent_proto_rawDesc), len(file_agent_agent_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   19,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_agent_agent_proto_goTypes,
		DependencyIndexes: file_agent_agent_proto_depIdxs,
		MessageInfos:      file_agent_agent_proto_msgTypes,
	}.Build()
	File_agent_agent_proto = out.File
	file_agent_agent_proto_goTypes = nil
	file_agent_agent_proto_depIdxs = nil
}
