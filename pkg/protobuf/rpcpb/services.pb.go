// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: rpcpb/services.proto

package rpcpb

import (
	builderpb "github.com/pygrum/monarch/pkg/protobuf/builderpb"
	clientpb "github.com/pygrum/monarch/pkg/protobuf/clientpb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogLevel int32

const (
	LogLevel_null         LogLevel = 0
	LogLevel_LevelDebug   LogLevel = 1
	LogLevel_LevelInfo    LogLevel = 2
	LogLevel_LevelSuccess LogLevel = 3
	LogLevel_LevelWarn    LogLevel = 4
	LogLevel_LevelError   LogLevel = 5
	LogLevel_LevelFatal   LogLevel = 6
)

// Enum value maps for LogLevel.
var (
	LogLevel_name = map[int32]string{
		0: "null",
		1: "LevelDebug",
		2: "LevelInfo",
		3: "LevelSuccess",
		4: "LevelWarn",
		5: "LevelError",
		6: "LevelFatal",
	}
	LogLevel_value = map[string]int32{
		"null":         0,
		"LevelDebug":   1,
		"LevelInfo":    2,
		"LevelSuccess": 3,
		"LevelWarn":    4,
		"LevelError":   5,
		"LevelFatal":   6,
	}
)

func (x LogLevel) Enum() *LogLevel {
	p := new(LogLevel)
	*p = x
	return p
}

func (x LogLevel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LogLevel) Descriptor() protoreflect.EnumDescriptor {
	return file_rpcpb_services_proto_enumTypes[0].Descriptor()
}

func (LogLevel) Type() protoreflect.EnumType {
	return &file_rpcpb_services_proto_enumTypes[0]
}

func (x LogLevel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LogLevel.Descriptor instead.
func (LogLevel) EnumDescriptor() ([]byte, []int) {
	return file_rpcpb_services_proto_rawDescGZIP(), []int{0}
}

type Notification struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogLevel LogLevel `protobuf:"varint,1,opt,name=log_level,json=logLevel,proto3,enum=rpcpb.LogLevel" json:"log_level,omitempty"`
	Msg      string   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *Notification) Reset() {
	*x = Notification{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcpb_services_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Notification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification) ProtoMessage() {}

func (x *Notification) ProtoReflect() protoreflect.Message {
	mi := &file_rpcpb_services_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification.ProtoReflect.Descriptor instead.
func (*Notification) Descriptor() ([]byte, []int) {
	return file_rpcpb_services_proto_rawDescGZIP(), []int{0}
}

func (x *Notification) GetLogLevel() LogLevel {
	if x != nil {
		return x.LogLevel
	}
	return LogLevel_null
}

func (x *Notification) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role string `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	From string `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	To   string `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Msg  string `protobuf:"bytes,4,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcpb_services_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_rpcpb_services_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_rpcpb_services_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

func (x *Message) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Message) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Message) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

var File_rpcpb_services_proto protoreflect.FileDescriptor

var file_rpcpb_services_proto_rawDesc = []byte{
	0x0a, 0x14, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x72, 0x70, 0x63, 0x70, 0x62, 0x1a, 0x17, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4e, 0x0a,
	0x0c, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a,
	0x09, 0x6c, 0x6f, 0x67, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0f, 0x2e, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x52, 0x08, 0x6c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x6d,
	0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x53, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d,
	0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f,
	0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d,
	0x73, 0x67, 0x2a, 0x74, 0x0a, 0x08, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x08,
	0x0a, 0x04, 0x6e, 0x75, 0x6c, 0x6c, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x44, 0x65, 0x62, 0x75, 0x67, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x49, 0x6e, 0x66, 0x6f, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x4c, 0x65, 0x76, 0x65, 0x6c,
	0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x4c, 0x65, 0x76,
	0x65, 0x6c, 0x57, 0x61, 0x72, 0x6e, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x05, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x46, 0x61, 0x74, 0x61, 0x6c, 0x10, 0x06, 0x32, 0xdc, 0x01, 0x0a, 0x07, 0x42, 0x75, 0x69,
	0x6c, 0x64, 0x65, 0x72, 0x12, 0x4d, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x73, 0x12, 0x1e, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x12, 0x42, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x19, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0a, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x17, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70,
	0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15,
	0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x32, 0xfa, 0x0e, 0x0a, 0x07, 0x4d, 0x6f, 0x6e, 0x61,
	0x72, 0x63, 0x68, 0x12, 0x37, 0x0a, 0x07, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x12, 0x17,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x06,
	0x41, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x16, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x73,
	0x22, 0x00, 0x12, 0x2e, 0x0a, 0x08, 0x4e, 0x65, 0x77, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x0f,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x1a,
	0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x00, 0x12, 0x35, 0x0a, 0x08, 0x52, 0x6d, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x16,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x08, 0x42, 0x75, 0x69,
	0x6c, 0x64, 0x65, 0x72, 0x73, 0x12, 0x18, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x12, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x65, 0x72, 0x73, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65,
	0x73, 0x12, 0x18, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x22,
	0x00, 0x12, 0x3e, 0x0a, 0x0b, 0x53, 0x61, 0x76, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65,
	0x12, 0x1c, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x61, 0x76, 0x65,
	0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x12, 0x44, 0x0a, 0x0b, 0x4c, 0x6f, 0x61, 0x64, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65,
	0x12, 0x1c, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x61, 0x76, 0x65,
	0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x0a, 0x52, 0x6d, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x18, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x00, 0x12, 0x3f, 0x0a, 0x07, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x19, 0x2e,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x65, 0x72, 0x70, 0x62, 0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x12, 0x38, 0x0a, 0x05, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x17, 0x2e, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x36, 0x0a,
	0x08, 0x45, 0x6e, 0x64, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x17, 0x2e, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x07, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x12, 0x18, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x72, 0x70, 0x63,
	0x70, 0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22,
	0x00, 0x30, 0x01, 0x12, 0x40, 0x0a, 0x09, 0x55, 0x6e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x12, 0x1a, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x55, 0x6e, 0x69, 0x6e,
	0x73, 0x74, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x72,
	0x70, 0x63, 0x70, 0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x22, 0x00, 0x30, 0x01, 0x12, 0x32, 0x0a, 0x08, 0x48, 0x74, 0x74, 0x70, 0x4f, 0x70, 0x65,
	0x6e, 0x12, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x1a, 0x13, 0x2e, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12, 0x2f, 0x0a, 0x09, 0x48, 0x74, 0x74,
	0x70, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x33, 0x0a, 0x09, 0x48, 0x74,
	0x74, 0x70, 0x73, 0x4f, 0x70, 0x65, 0x6e, 0x12, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x72, 0x70, 0x63, 0x70, 0x62,
	0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12,
	0x30, 0x0a, 0x0a, 0x48, 0x74, 0x74, 0x70, 0x73, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x0f, 0x2e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0f,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x12, 0x31, 0x0a, 0x07, 0x54, 0x63, 0x70, 0x4f, 0x70, 0x65, 0x6e, 0x12, 0x0f, 0x2e, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e,
	0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x08, 0x54, 0x63, 0x70, 0x43, 0x6c, 0x6f, 0x73, 0x65,
	0x12, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x00, 0x12, 0x3b, 0x0a, 0x08, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x19, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22,
	0x00, 0x12, 0x39, 0x0a, 0x09, 0x52, 0x6d, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x19,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0b,
	0x4c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0b,
	0x46, 0x72, 0x65, 0x65, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x46, 0x72, 0x65, 0x65, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x4a, 0x0a, 0x08,
	0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12, 0x1e, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x65, 0x72, 0x70, 0x62, 0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x65, 0x72, 0x70, 0x62, 0x2e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x04, 0x53, 0x65, 0x6e, 0x64,
	0x12, 0x15, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x48, 0x54, 0x54, 0x50,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x2f, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x67, 0x65, 0x56, 0x69, 0x65, 0x77, 0x12, 0x0f,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a,
	0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x67, 0x65,
	0x22, 0x00, 0x12, 0x3c, 0x0a, 0x08, 0x53, 0x74, 0x61, 0x67, 0x65, 0x41, 0x64, 0x64, 0x12, 0x19,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x67, 0x65, 0x41,
	0x64, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x72, 0x70, 0x63, 0x70,
	0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00,
	0x12, 0x40, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x67, 0x65, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x12, 0x1b,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x67, 0x65, 0x4c,
	0x6f, 0x63, 0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x72, 0x70,
	0x63, 0x70, 0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x22, 0x00, 0x12, 0x36, 0x0a, 0x07, 0x55, 0x6e, 0x73, 0x74, 0x61, 0x67, 0x65, 0x12, 0x18, 0x2e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x55, 0x6e, 0x73, 0x74, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x32, 0x0a, 0x06, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x12, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x30, 0x01, 0x12, 0x32,
	0x0a, 0x0b, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x0f, 0x2e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0e,
	0x2e, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00,
	0x30, 0x01, 0x12, 0x30, 0x0a, 0x0b, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x0e, 0x2e, 0x72, 0x70, 0x63, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x1a, 0x0f, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x00, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x70, 0x79, 0x67, 0x72, 0x75, 0x6d, 0x2f, 0x6d, 0x6f, 0x6e, 0x61, 0x72, 0x63,
	0x68, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x72,
	0x70, 0x63, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpcpb_services_proto_rawDescOnce sync.Once
	file_rpcpb_services_proto_rawDescData = file_rpcpb_services_proto_rawDesc
)

func file_rpcpb_services_proto_rawDescGZIP() []byte {
	file_rpcpb_services_proto_rawDescOnce.Do(func() {
		file_rpcpb_services_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpcpb_services_proto_rawDescData)
	})
	return file_rpcpb_services_proto_rawDescData
}

var file_rpcpb_services_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_rpcpb_services_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rpcpb_services_proto_goTypes = []interface{}{
	(LogLevel)(0),                         // 0: rpcpb.LogLevel
	(*Notification)(nil),                  // 1: rpcpb.Notification
	(*Message)(nil),                       // 2: rpcpb.Message
	(*builderpb.DescriptionsRequest)(nil), // 3: builderpb.DescriptionsRequest
	(*builderpb.OptionsRequest)(nil),      // 4: builderpb.OptionsRequest
	(*builderpb.BuildRequest)(nil),        // 5: builderpb.BuildRequest
	(*clientpb.PlayerRequest)(nil),        // 6: clientpb.PlayerRequest
	(*clientpb.AgentRequest)(nil),         // 7: clientpb.AgentRequest
	(*clientpb.Agent)(nil),                // 8: clientpb.Agent
	(*clientpb.BuilderRequest)(nil),       // 9: clientpb.BuilderRequest
	(*clientpb.ProfileRequest)(nil),       // 10: clientpb.ProfileRequest
	(*clientpb.SaveProfileRequest)(nil),   // 11: clientpb.SaveProfileRequest
	(*clientpb.InstallRequest)(nil),       // 12: clientpb.InstallRequest
	(*clientpb.UninstallRequest)(nil),     // 13: clientpb.UninstallRequest
	(*clientpb.Empty)(nil),                // 14: clientpb.Empty
	(*clientpb.SessionsRequest)(nil),      // 15: clientpb.SessionsRequest
	(*clientpb.LockSessionRequest)(nil),   // 16: clientpb.LockSessionRequest
	(*clientpb.FreeSessionRequest)(nil),   // 17: clientpb.FreeSessionRequest
	(*clientpb.HTTPRequest)(nil),          // 18: clientpb.HTTPRequest
	(*clientpb.StageAddRequest)(nil),      // 19: clientpb.StageAddRequest
	(*clientpb.StageLocalRequest)(nil),    // 20: clientpb.StageLocalRequest
	(*clientpb.UnstageRequest)(nil),       // 21: clientpb.UnstageRequest
	(*builderpb.DescriptionsReply)(nil),   // 22: builderpb.DescriptionsReply
	(*builderpb.OptionsReply)(nil),        // 23: builderpb.OptionsReply
	(*builderpb.BuildReply)(nil),          // 24: builderpb.BuildReply
	(*clientpb.Players)(nil),              // 25: clientpb.Players
	(*clientpb.Agents)(nil),               // 26: clientpb.Agents
	(*clientpb.Builders)(nil),             // 27: clientpb.Builders
	(*clientpb.Profiles)(nil),             // 28: clientpb.Profiles
	(*clientpb.ProfileData)(nil),          // 29: clientpb.ProfileData
	(*clientpb.BuildReply)(nil),           // 30: clientpb.BuildReply
	(*clientpb.Sessions)(nil),             // 31: clientpb.Sessions
	(*clientpb.HTTPResponse)(nil),         // 32: clientpb.HTTPResponse
	(*clientpb.Stage)(nil),                // 33: clientpb.Stage
}
var file_rpcpb_services_proto_depIdxs = []int32{
	0,  // 0: rpcpb.Notification.log_level:type_name -> rpcpb.LogLevel
	3,  // 1: rpcpb.Builder.GetCommands:input_type -> builderpb.DescriptionsRequest
	4,  // 2: rpcpb.Builder.GetOptions:input_type -> builderpb.OptionsRequest
	5,  // 3: rpcpb.Builder.BuildAgent:input_type -> builderpb.BuildRequest
	6,  // 4: rpcpb.Monarch.Players:input_type -> clientpb.PlayerRequest
	7,  // 5: rpcpb.Monarch.Agents:input_type -> clientpb.AgentRequest
	8,  // 6: rpcpb.Monarch.NewAgent:input_type -> clientpb.Agent
	7,  // 7: rpcpb.Monarch.RmAgents:input_type -> clientpb.AgentRequest
	9,  // 8: rpcpb.Monarch.Builders:input_type -> clientpb.BuilderRequest
	10, // 9: rpcpb.Monarch.Profiles:input_type -> clientpb.ProfileRequest
	11, // 10: rpcpb.Monarch.SaveProfile:input_type -> clientpb.SaveProfileRequest
	11, // 11: rpcpb.Monarch.LoadProfile:input_type -> clientpb.SaveProfileRequest
	10, // 12: rpcpb.Monarch.RmProfiles:input_type -> clientpb.ProfileRequest
	4,  // 13: rpcpb.Monarch.Options:input_type -> builderpb.OptionsRequest
	5,  // 14: rpcpb.Monarch.Build:input_type -> builderpb.BuildRequest
	5,  // 15: rpcpb.Monarch.EndBuild:input_type -> builderpb.BuildRequest
	12, // 16: rpcpb.Monarch.Install:input_type -> clientpb.InstallRequest
	13, // 17: rpcpb.Monarch.Uninstall:input_type -> clientpb.UninstallRequest
	14, // 18: rpcpb.Monarch.HttpOpen:input_type -> clientpb.Empty
	14, // 19: rpcpb.Monarch.HttpClose:input_type -> clientpb.Empty
	14, // 20: rpcpb.Monarch.HttpsOpen:input_type -> clientpb.Empty
	14, // 21: rpcpb.Monarch.HttpsClose:input_type -> clientpb.Empty
	14, // 22: rpcpb.Monarch.TcpOpen:input_type -> clientpb.Empty
	14, // 23: rpcpb.Monarch.TcpClose:input_type -> clientpb.Empty
	15, // 24: rpcpb.Monarch.Sessions:input_type -> clientpb.SessionsRequest
	15, // 25: rpcpb.Monarch.RmSession:input_type -> clientpb.SessionsRequest
	16, // 26: rpcpb.Monarch.LockSession:input_type -> clientpb.LockSessionRequest
	17, // 27: rpcpb.Monarch.FreeSession:input_type -> clientpb.FreeSessionRequest
	3,  // 28: rpcpb.Monarch.Commands:input_type -> builderpb.DescriptionsRequest
	18, // 29: rpcpb.Monarch.Send:input_type -> clientpb.HTTPRequest
	14, // 30: rpcpb.Monarch.StageView:input_type -> clientpb.Empty
	19, // 31: rpcpb.Monarch.StageAdd:input_type -> clientpb.StageAddRequest
	20, // 32: rpcpb.Monarch.StageLocal:input_type -> clientpb.StageLocalRequest
	21, // 33: rpcpb.Monarch.Unstage:input_type -> clientpb.UnstageRequest
	14, // 34: rpcpb.Monarch.Notify:input_type -> clientpb.Empty
	14, // 35: rpcpb.Monarch.GetMessages:input_type -> clientpb.Empty
	2,  // 36: rpcpb.Monarch.SendMessage:input_type -> rpcpb.Message
	22, // 37: rpcpb.Builder.GetCommands:output_type -> builderpb.DescriptionsReply
	23, // 38: rpcpb.Builder.GetOptions:output_type -> builderpb.OptionsReply
	24, // 39: rpcpb.Builder.BuildAgent:output_type -> builderpb.BuildReply
	25, // 40: rpcpb.Monarch.Players:output_type -> clientpb.Players
	26, // 41: rpcpb.Monarch.Agents:output_type -> clientpb.Agents
	14, // 42: rpcpb.Monarch.NewAgent:output_type -> clientpb.Empty
	14, // 43: rpcpb.Monarch.RmAgents:output_type -> clientpb.Empty
	27, // 44: rpcpb.Monarch.Builders:output_type -> clientpb.Builders
	28, // 45: rpcpb.Monarch.Profiles:output_type -> clientpb.Profiles
	14, // 46: rpcpb.Monarch.SaveProfile:output_type -> clientpb.Empty
	29, // 47: rpcpb.Monarch.LoadProfile:output_type -> clientpb.ProfileData
	14, // 48: rpcpb.Monarch.RmProfiles:output_type -> clientpb.Empty
	23, // 49: rpcpb.Monarch.Options:output_type -> builderpb.OptionsReply
	30, // 50: rpcpb.Monarch.Build:output_type -> clientpb.BuildReply
	14, // 51: rpcpb.Monarch.EndBuild:output_type -> clientpb.Empty
	1,  // 52: rpcpb.Monarch.Install:output_type -> rpcpb.Notification
	1,  // 53: rpcpb.Monarch.Uninstall:output_type -> rpcpb.Notification
	1,  // 54: rpcpb.Monarch.HttpOpen:output_type -> rpcpb.Notification
	14, // 55: rpcpb.Monarch.HttpClose:output_type -> clientpb.Empty
	1,  // 56: rpcpb.Monarch.HttpsOpen:output_type -> rpcpb.Notification
	14, // 57: rpcpb.Monarch.HttpsClose:output_type -> clientpb.Empty
	1,  // 58: rpcpb.Monarch.TcpOpen:output_type -> rpcpb.Notification
	14, // 59: rpcpb.Monarch.TcpClose:output_type -> clientpb.Empty
	31, // 60: rpcpb.Monarch.Sessions:output_type -> clientpb.Sessions
	14, // 61: rpcpb.Monarch.RmSession:output_type -> clientpb.Empty
	14, // 62: rpcpb.Monarch.LockSession:output_type -> clientpb.Empty
	14, // 63: rpcpb.Monarch.FreeSession:output_type -> clientpb.Empty
	22, // 64: rpcpb.Monarch.Commands:output_type -> builderpb.DescriptionsReply
	32, // 65: rpcpb.Monarch.Send:output_type -> clientpb.HTTPResponse
	33, // 66: rpcpb.Monarch.StageView:output_type -> clientpb.Stage
	1,  // 67: rpcpb.Monarch.StageAdd:output_type -> rpcpb.Notification
	1,  // 68: rpcpb.Monarch.StageLocal:output_type -> rpcpb.Notification
	14, // 69: rpcpb.Monarch.Unstage:output_type -> clientpb.Empty
	1,  // 70: rpcpb.Monarch.Notify:output_type -> rpcpb.Notification
	2,  // 71: rpcpb.Monarch.GetMessages:output_type -> rpcpb.Message
	14, // 72: rpcpb.Monarch.SendMessage:output_type -> clientpb.Empty
	37, // [37:73] is the sub-list for method output_type
	1,  // [1:37] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_rpcpb_services_proto_init() }
func file_rpcpb_services_proto_init() {
	if File_rpcpb_services_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rpcpb_services_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Notification); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rpcpb_services_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rpcpb_services_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_rpcpb_services_proto_goTypes,
		DependencyIndexes: file_rpcpb_services_proto_depIdxs,
		EnumInfos:         file_rpcpb_services_proto_enumTypes,
		MessageInfos:      file_rpcpb_services_proto_msgTypes,
	}.Build()
	File_rpcpb_services_proto = out.File
	file_rpcpb_services_proto_rawDesc = nil
	file_rpcpb_services_proto_goTypes = nil
	file_rpcpb_services_proto_depIdxs = nil
}
