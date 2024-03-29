// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: examples.proto

// The protocol buffer FQDN. Used by the goomerang
// internals to distinguish among messages.

package protos

import (
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

// MessageV1 represent the first version
// of a message.
type MessageV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *MessageV1) Reset() {
	*x = MessageV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examples_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageV1) ProtoMessage() {}

func (x *MessageV1) ProtoReflect() protoreflect.Message {
	mi := &file_examples_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageV1.ProtoReflect.Descriptor instead.
func (*MessageV1) Descriptor() ([]byte, []int) {
	return file_examples_proto_rawDescGZIP(), []int{0}
}

func (x *MessageV1) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// ReplyMsgV1 represent the first version of
// a reply message.
type ReplyV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ReplyV1) Reset() {
	*x = ReplyV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examples_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReplyV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplyV1) ProtoMessage() {}

func (x *ReplyV1) ProtoReflect() protoreflect.Message {
	mi := &file_examples_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplyV1.ProtoReflect.Descriptor instead.
func (*ReplyV1) Descriptor() ([]byte, []int) {
	return file_examples_proto_rawDescGZIP(), []int{1}
}

func (x *ReplyV1) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// BadReplyV1 represent the first version of a
// bad reply by the server.
type BadReplyV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *BadReplyV1) Reset() {
	*x = BadReplyV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examples_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BadReplyV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BadReplyV1) ProtoMessage() {}

func (x *BadReplyV1) ProtoReflect() protoreflect.Message {
	mi := &file_examples_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BadReplyV1.ProtoReflect.Descriptor instead.
func (*BadReplyV1) Descriptor() ([]byte, []int) {
	return file_examples_proto_rawDescGZIP(), []int{2}
}

func (x *BadReplyV1) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_examples_proto protoreflect.FileDescriptor

var file_examples_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x11, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2e, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x22, 0x25, 0x0a, 0x09, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x56, 0x31,
	0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x23, 0x0a, 0x07, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x56, 0x31, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22,
	0x26, 0x0a, 0x0a, 0x42, 0x61, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x56, 0x31, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x6f, 0x2e, 0x65, 0x6c,
	0x6f, 0x79, 0x6c, 0x70, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61,
	0x6e, 0x67, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_examples_proto_rawDescOnce sync.Once
	file_examples_proto_rawDescData = file_examples_proto_rawDesc
)

func file_examples_proto_rawDescGZIP() []byte {
	file_examples_proto_rawDescOnce.Do(func() {
		file_examples_proto_rawDescData = protoimpl.X.CompressGZIP(file_examples_proto_rawDescData)
	})
	return file_examples_proto_rawDescData
}

var file_examples_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_examples_proto_goTypes = []interface{}{
	(*MessageV1)(nil),  // 0: goomerang.example.MessageV1
	(*ReplyV1)(nil),    // 1: goomerang.example.ReplyV1
	(*BadReplyV1)(nil), // 2: goomerang.example.BadReplyV1
}
var file_examples_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_examples_proto_init() }
func file_examples_proto_init() {
	if File_examples_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_examples_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageV1); i {
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
		file_examples_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReplyV1); i {
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
		file_examples_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BadReplyV1); i {
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
			RawDescriptor: file_examples_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_examples_proto_goTypes,
		DependencyIndexes: file_examples_proto_depIdxs,
		MessageInfos:      file_examples_proto_msgTypes,
	}.Build()
	File_examples_proto = out.File
	file_examples_proto_rawDesc = nil
	file_examples_proto_goTypes = nil
	file_examples_proto_depIdxs = nil
}
