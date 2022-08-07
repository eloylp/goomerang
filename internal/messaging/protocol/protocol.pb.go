// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: internal/messaging/protocol.proto

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

	Uuid        string                 `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Kind        string                 `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	PayloadSize int64                  `protobuf:"varint,3,opt,name=payloadSize,proto3" json:"payloadSize,omitempty"`
	IsSync      bool                   `protobuf:"varint,4,opt,name=is_sync,json=isSync,proto3" json:"is_sync,omitempty"`
	Creation    *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=creation,proto3" json:"creation,omitempty"`
	Headers     map[string]string      `protobuf:"bytes,6,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Payload     []byte                 `protobuf:"bytes,7,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *Frame) Reset() {
	*x = Frame{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_messaging_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Frame) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Frame) ProtoMessage() {}

func (x *Frame) ProtoReflect() protoreflect.Message {
	mi := &file_internal_messaging_protocol_proto_msgTypes[0]
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
	return file_internal_messaging_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *Frame) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Frame) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Frame) GetPayloadSize() int64 {
	if x != nil {
		return x.PayloadSize
	}
	return 0
}

func (x *Frame) GetIsSync() bool {
	if x != nil {
		return x.IsSync
	}
	return false
}

func (x *Frame) GetCreation() *timestamppb.Timestamp {
	if x != nil {
		return x.Creation
	}
	return nil
}

func (x *Frame) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Frame) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type BroadcastCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kind    string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Message []byte `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *BroadcastCmd) Reset() {
	*x = BroadcastCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_messaging_protocol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BroadcastCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BroadcastCmd) ProtoMessage() {}

func (x *BroadcastCmd) ProtoReflect() protoreflect.Message {
	mi := &file_internal_messaging_protocol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BroadcastCmd.ProtoReflect.Descriptor instead.
func (*BroadcastCmd) Descriptor() ([]byte, []int) {
	return file_internal_messaging_protocol_proto_rawDescGZIP(), []int{1}
}

func (x *BroadcastCmd) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *BroadcastCmd) GetMessage() []byte {
	if x != nil {
		return x.Message
	}
	return nil
}

type SubscribeCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
}

func (x *SubscribeCmd) Reset() {
	*x = SubscribeCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_messaging_protocol_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeCmd) ProtoMessage() {}

func (x *SubscribeCmd) ProtoReflect() protoreflect.Message {
	mi := &file_internal_messaging_protocol_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeCmd.ProtoReflect.Descriptor instead.
func (*SubscribeCmd) Descriptor() ([]byte, []int) {
	return file_internal_messaging_protocol_proto_rawDescGZIP(), []int{2}
}

func (x *SubscribeCmd) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

type PublishCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic   string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Kind    string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Message []byte `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *PublishCmd) Reset() {
	*x = PublishCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_messaging_protocol_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishCmd) ProtoMessage() {}

func (x *PublishCmd) ProtoReflect() protoreflect.Message {
	mi := &file_internal_messaging_protocol_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishCmd.ProtoReflect.Descriptor instead.
func (*PublishCmd) Descriptor() ([]byte, []int) {
	return file_internal_messaging_protocol_proto_rawDescGZIP(), []int{3}
}

func (x *PublishCmd) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *PublishCmd) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *PublishCmd) GetMessage() []byte {
	if x != nil {
		return x.Message
	}
	return nil
}

type UnsubscribeCmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
}

func (x *UnsubscribeCmd) Reset() {
	*x = UnsubscribeCmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_messaging_protocol_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnsubscribeCmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsubscribeCmd) ProtoMessage() {}

func (x *UnsubscribeCmd) ProtoReflect() protoreflect.Message {
	mi := &file_internal_messaging_protocol_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsubscribeCmd.ProtoReflect.Descriptor instead.
func (*UnsubscribeCmd) Descriptor() ([]byte, []int) {
	return file_internal_messaging_protocol_proto_rawDescGZIP(), []int{4}
}

func (x *UnsubscribeCmd) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

var File_internal_messaging_protocol_proto protoreflect.FileDescriptor

var file_internal_messaging_protocol_proto_rawDesc = []byte{
	0x0a, 0x21, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x12, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xba, 0x02, 0x0a, 0x05, 0x46, 0x72, 0x61,
	0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0b, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x17, 0x0a, 0x07,
	0x69, 0x73, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69,
	0x73, 0x53, 0x79, 0x6e, 0x63, 0x12, 0x36, 0x0a, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x40, 0x0a,
	0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26,
	0x2e, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x1a, 0x3a, 0x0a, 0x0c, 0x48, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x3c, 0x0a, 0x0c, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61,
	0x73, 0x74, 0x43, 0x6d, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0x24, 0x0a, 0x0c, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65,
	0x43, 0x6d, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x22, 0x50, 0x0a, 0x0a, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x43, 0x6d, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x12, 0x0a,
	0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x26, 0x0a, 0x0e, 0x55,
	0x6e, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x43, 0x6d, 0x64, 0x12, 0x14, 0x0a,
	0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x6f, 0x2e, 0x65, 0x6c, 0x6f, 0x79, 0x6c, 0x70,
	0x2e, 0x64, 0x65, 0x76, 0x2f, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_messaging_protocol_proto_rawDescOnce sync.Once
	file_internal_messaging_protocol_proto_rawDescData = file_internal_messaging_protocol_proto_rawDesc
)

func file_internal_messaging_protocol_proto_rawDescGZIP() []byte {
	file_internal_messaging_protocol_proto_rawDescOnce.Do(func() {
		file_internal_messaging_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_messaging_protocol_proto_rawDescData)
	})
	return file_internal_messaging_protocol_proto_rawDescData
}

var file_internal_messaging_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_internal_messaging_protocol_proto_goTypes = []interface{}{
	(*Frame)(nil),                 // 0: goomerang.protocol.Frame
	(*BroadcastCmd)(nil),          // 1: goomerang.protocol.BroadcastCmd
	(*SubscribeCmd)(nil),          // 2: goomerang.protocol.SubscribeCmd
	(*PublishCmd)(nil),            // 3: goomerang.protocol.PublishCmd
	(*UnsubscribeCmd)(nil),        // 4: goomerang.protocol.UnsubscribeCmd
	nil,                           // 5: goomerang.protocol.Frame.HeadersEntry
	(*timestamppb.Timestamp)(nil), // 6: google.protobuf.Timestamp
}
var file_internal_messaging_protocol_proto_depIdxs = []int32{
	6, // 0: goomerang.protocol.Frame.creation:type_name -> google.protobuf.Timestamp
	5, // 1: goomerang.protocol.Frame.headers:type_name -> goomerang.protocol.Frame.HeadersEntry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_internal_messaging_protocol_proto_init() }
func file_internal_messaging_protocol_proto_init() {
	if File_internal_messaging_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_messaging_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_messaging_protocol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BroadcastCmd); i {
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
		file_internal_messaging_protocol_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeCmd); i {
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
		file_internal_messaging_protocol_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublishCmd); i {
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
		file_internal_messaging_protocol_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnsubscribeCmd); i {
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
			RawDescriptor: file_internal_messaging_protocol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_messaging_protocol_proto_goTypes,
		DependencyIndexes: file_internal_messaging_protocol_proto_depIdxs,
		MessageInfos:      file_internal_messaging_protocol_proto_msgTypes,
	}.Build()
	File_internal_messaging_protocol_proto = out.File
	file_internal_messaging_protocol_proto_rawDesc = nil
	file_internal_messaging_protocol_proto_goTypes = nil
	file_internal_messaging_protocol_proto_depIdxs = nil
}
