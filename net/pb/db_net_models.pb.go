// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.12.4
// source: db_net_models.proto

package pb

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type DBAddrInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LastSeen *timestamp.Timestamp `protobuf:"bytes,1,opt,name=last_seen,json=lastSeen,proto3" json:"last_seen,omitempty"`
	Addrs    [][]byte             `protobuf:"bytes,2,rep,name=addrs,proto3" json:"addrs,omitempty"`
}

func (x *DBAddrInfo) Reset() {
	*x = DBAddrInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_db_net_models_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DBAddrInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DBAddrInfo) ProtoMessage() {}

func (x *DBAddrInfo) ProtoReflect() protoreflect.Message {
	mi := &file_db_net_models_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DBAddrInfo.ProtoReflect.Descriptor instead.
func (*DBAddrInfo) Descriptor() ([]byte, []int) {
	return file_db_net_models_proto_rawDescGZIP(), []int{0}
}

func (x *DBAddrInfo) GetLastSeen() *timestamp.Timestamp {
	if x != nil {
		return x.LastSeen
	}
	return nil
}

func (x *DBAddrInfo) GetAddrs() [][]byte {
	if x != nil {
		return x.Addrs
	}
	return nil
}

var File_db_net_models_proto protoreflect.FileDescriptor

var file_db_net_models_proto_rawDesc = []byte{
	0x0a, 0x13, 0x64, 0x62, 0x5f, 0x6e, 0x65, 0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5b, 0x0a, 0x0a, 0x44, 0x42, 0x41, 0x64, 0x64, 0x72,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x37, 0x0a, 0x09, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x73, 0x65, 0x65,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x08, 0x6c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x12, 0x14, 0x0a,
	0x05, 0x61, 0x64, 0x64, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x05, 0x61, 0x64,
	0x64, 0x72, 0x73, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_db_net_models_proto_rawDescOnce sync.Once
	file_db_net_models_proto_rawDescData = file_db_net_models_proto_rawDesc
)

func file_db_net_models_proto_rawDescGZIP() []byte {
	file_db_net_models_proto_rawDescOnce.Do(func() {
		file_db_net_models_proto_rawDescData = protoimpl.X.CompressGZIP(file_db_net_models_proto_rawDescData)
	})
	return file_db_net_models_proto_rawDescData
}

var file_db_net_models_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_db_net_models_proto_goTypes = []interface{}{
	(*DBAddrInfo)(nil),          // 0: DBAddrInfo
	(*timestamp.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_db_net_models_proto_depIdxs = []int32{
	1, // 0: DBAddrInfo.last_seen:type_name -> google.protobuf.Timestamp
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_db_net_models_proto_init() }
func file_db_net_models_proto_init() {
	if File_db_net_models_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_db_net_models_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DBAddrInfo); i {
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
			RawDescriptor: file_db_net_models_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_db_net_models_proto_goTypes,
		DependencyIndexes: file_db_net_models_proto_depIdxs,
		MessageInfos:      file_db_net_models_proto_msgTypes,
	}.Build()
	File_db_net_models_proto = out.File
	file_db_net_models_proto_rawDesc = nil
	file_db_net_models_proto_goTypes = nil
	file_db_net_models_proto_depIdxs = nil
}
