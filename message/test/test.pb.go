// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1-devel
// 	protoc        v3.14.0
// source: message/test.proto

package test

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

type PingPong struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *PingPong) Reset() {
	*x = PingPong{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_test_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingPong) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingPong) ProtoMessage() {}

func (x *PingPong) ProtoReflect() protoreflect.Message {
	mi := &file_message_test_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingPong.ProtoReflect.Descriptor instead.
func (*PingPong) Descriptor() ([]byte, []int) {
	return file_message_test_proto_rawDescGZIP(), []int{0}
}

func (x *PingPong) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type GreetV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *GreetV1) Reset() {
	*x = GreetV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_test_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GreetV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreetV1) ProtoMessage() {}

func (x *GreetV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_test_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GreetV1.ProtoReflect.Descriptor instead.
func (*GreetV1) Descriptor() ([]byte, []int) {
	return file_message_test_proto_rawDescGZIP(), []int{1}
}

func (x *GreetV1) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_message_test_proto protoreflect.FileDescriptor

var file_message_test_proto_rawDesc = []byte{
	0x0a, 0x12, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x67, 0x6f, 0x6f, 0x6d, 0x65, 0x72, 0x61, 0x6e, 0x67, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x22, 0x24, 0x0a, 0x08, 0x50, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6e, 0x67,
	0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x23, 0x0a, 0x07, 0x47, 0x72,
	0x65, 0x65, 0x74, 0x56, 0x31, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42,
	0x07, 0x5a, 0x05, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_test_proto_rawDescOnce sync.Once
	file_message_test_proto_rawDescData = file_message_test_proto_rawDesc
)

func file_message_test_proto_rawDescGZIP() []byte {
	file_message_test_proto_rawDescOnce.Do(func() {
		file_message_test_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_test_proto_rawDescData)
	})
	return file_message_test_proto_rawDescData
}

var file_message_test_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_message_test_proto_goTypes = []interface{}{
	(*PingPong)(nil), // 0: goomerang.test.PingPong
	(*GreetV1)(nil),  // 1: goomerang.test.GreetV1
}
var file_message_test_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_message_test_proto_init() }
func file_message_test_proto_init() {
	if File_message_test_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_message_test_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingPong); i {
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
		file_message_test_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GreetV1); i {
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
			RawDescriptor: file_message_test_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_test_proto_goTypes,
		DependencyIndexes: file_message_test_proto_depIdxs,
		MessageInfos:      file_message_test_proto_msgTypes,
	}.Build()
	File_message_test_proto = out.File
	file_message_test_proto_rawDesc = nil
	file_message_test_proto_goTypes = nil
	file_message_test_proto_depIdxs = nil
}