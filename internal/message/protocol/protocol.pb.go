// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: internal/message/protocol.proto

package protocol

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Frame struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid     string                 `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Type     string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	IsRpc    bool                   `protobuf:"varint,3,opt,name=is_rpc,json=isRpc,proto3" json:"is_rpc,omitempty"`
	Creation *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=creation,proto3" json:"creation,omitempty"`
	Payload  []byte                 `protobuf:"bytes,5,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *Frame) Reset() {
	*x = Frame{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_message_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Frame) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Frame) ProtoMessage() {}

func (x *Frame) ProtoReflect() protoreflect.Message {
	mi := &file_internal_message_protocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Frame.ProtoReflect.Descriptor instead.
func (*Frame) Descriptor() ([]byte, []int) {
	return file_internal_message_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *Frame) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Frame) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Frame) GetIsRpc() bool {
	if x != nil {
		return x.IsRpc
	}
	return false
}

func (x *Frame) GetCreation() *timestamppb.Timestamp {
	if x != nil {
		return x.Creation
	}
	return nil
}

func (x *Frame) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

var File_internal_message_protocol_proto protoreflect.FileDescriptor

var file_internal_message_protocol_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x12, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x98, 0x01, 0x0a, 0x05, 0x46, 0x72, 0x61, 0x6d, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x75, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x69, 0x73, 0x5f, 0x72,
	0x70, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x69, 0x73, 0x52, 0x70, 0x63, 0x12,
	0x36, 0x0a, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x6f, 0x2e, 0x65, 0x6c, 0x6f, 0x79, 0x6c, 0x70, 0x2e, 0x64,
	0x65, 0x76, 0x2f, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_message_protocol_proto_rawDescOnce sync.Once
	file_internal_message_protocol_proto_rawDescData = file_internal_message_protocol_proto_rawDesc
)

func file_internal_message_protocol_proto_rawDescGZIP() []byte {
	file_internal_message_protocol_proto_rawDescOnce.Do(func() {
		file_internal_message_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_message_protocol_proto_rawDescData)
	})
	return file_internal_message_protocol_proto_rawDescData
}

var file_internal_message_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_internal_message_protocol_proto_goTypes = []interface{}{
	(*Frame)(nil),                 // 0: goomerang.protocol.Frame
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_internal_message_protocol_proto_depIdxs = []int32{
	1, // 0: goomerang.protocol.Frame.creation:type_name -> google.protobuf.Timestamp
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_message_protocol_proto_init() }
func file_internal_message_protocol_proto_init() {
	if File_internal_message_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_message_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Frame); i {
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
			RawDescriptor: file_internal_message_protocol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_message_protocol_proto_goTypes,
		DependencyIndexes: file_internal_message_protocol_proto_depIdxs,
		MessageInfos:      file_internal_message_protocol_proto_msgTypes,
	}.Build()
	File_internal_message_protocol_proto = out.File
	file_internal_message_protocol_proto_rawDesc = nil
	file_internal_message_protocol_proto_goTypes = nil
	file_internal_message_protocol_proto_depIdxs = nil
}
