// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: machinetalk/protobuf/value.proto

// see README.msgid
// msgid base: 1500

package machinetalk_protobuf_go

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

type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type *ValueType `protobuf:"varint,10,req,name=type,enum=machinetalk.ValueType" json:"type,omitempty"` // actual values
	// scalars
	Halbit   *bool    `protobuf:"varint,100,opt,name=halbit" json:"halbit,omitempty"`
	Halfloat *float64 `protobuf:"fixed64,101,opt,name=halfloat" json:"halfloat,omitempty"`
	Hals32   *int32   `protobuf:"fixed32,102,opt,name=hals32" json:"hals32,omitempty"`
	Halu32   *uint32  `protobuf:"fixed32,103,opt,name=halu32" json:"halu32,omitempty"`
	VBytes   []byte   `protobuf:"bytes,120,opt,name=v_bytes,json=vBytes" json:"v_bytes,omitempty"`
	VInt32   *int32   `protobuf:"fixed32,130,opt,name=v_int32,json=vInt32" json:"v_int32,omitempty"`
	VInt64   *int64   `protobuf:"fixed64,140,opt,name=v_int64,json=vInt64" json:"v_int64,omitempty"`
	VUint32  *uint32  `protobuf:"fixed32,150,opt,name=v_uint32,json=vUint32" json:"v_uint32,omitempty"`
	VUint64  *uint64  `protobuf:"fixed64,160,opt,name=v_uint64,json=vUint64" json:"v_uint64,omitempty"`
	VDouble  *float64 `protobuf:"fixed64,170,opt,name=v_double,json=vDouble" json:"v_double,omitempty"`
	VString  *string  `protobuf:"bytes,180,opt,name=v_string,json=vString" json:"v_string,omitempty"`
	VBool    *bool    `protobuf:"varint,190,opt,name=v_bool,json=vBool" json:"v_bool,omitempty"`
	// compound types
	Carte *PmCartesian `protobuf:"bytes,200,opt,name=carte" json:"carte,omitempty"`
	Pose  *EmcPose     `protobuf:"bytes,220,opt,name=pose" json:"pose,omitempty"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_value_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_value_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_value_proto_rawDescGZIP(), []int{0}
}

func (x *Value) GetType() ValueType {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return ValueType_HAL_BIT
}

func (x *Value) GetHalbit() bool {
	if x != nil && x.Halbit != nil {
		return *x.Halbit
	}
	return false
}

func (x *Value) GetHalfloat() float64 {
	if x != nil && x.Halfloat != nil {
		return *x.Halfloat
	}
	return 0
}

func (x *Value) GetHals32() int32 {
	if x != nil && x.Hals32 != nil {
		return *x.Hals32
	}
	return 0
}

func (x *Value) GetHalu32() uint32 {
	if x != nil && x.Halu32 != nil {
		return *x.Halu32
	}
	return 0
}

func (x *Value) GetVBytes() []byte {
	if x != nil {
		return x.VBytes
	}
	return nil
}

func (x *Value) GetVInt32() int32 {
	if x != nil && x.VInt32 != nil {
		return *x.VInt32
	}
	return 0
}

func (x *Value) GetVInt64() int64 {
	if x != nil && x.VInt64 != nil {
		return *x.VInt64
	}
	return 0
}

func (x *Value) GetVUint32() uint32 {
	if x != nil && x.VUint32 != nil {
		return *x.VUint32
	}
	return 0
}

func (x *Value) GetVUint64() uint64 {
	if x != nil && x.VUint64 != nil {
		return *x.VUint64
	}
	return 0
}

func (x *Value) GetVDouble() float64 {
	if x != nil && x.VDouble != nil {
		return *x.VDouble
	}
	return 0
}

func (x *Value) GetVString() string {
	if x != nil && x.VString != nil {
		return *x.VString
	}
	return ""
}

func (x *Value) GetVBool() bool {
	if x != nil && x.VBool != nil {
		return *x.VBool
	}
	return false
}

func (x *Value) GetCarte() *PmCartesian {
	if x != nil {
		return x.Carte
	}
	return nil
}

func (x *Value) GetPose() *EmcPose {
	if x != nil {
		return x.Pose
	}
	return nil
}

var File_machinetalk_protobuf_value_proto protoreflect.FileDescriptor

var file_machinetalk_protobuf_value_proto_rawDesc = []byte{
	0x0a, 0x20, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0b, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x1a,
	0x21, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x61, 0x6e, 0x6f, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x23, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x63, 0x63, 0x6c, 0x61, 0x73,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65,
	0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd7, 0x03, 0x0a, 0x05, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x2a, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0a, 0x20, 0x02, 0x28,
	0x0e, 0x32, 0x16, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2e,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x68, 0x61, 0x6c, 0x62, 0x69, 0x74, 0x18, 0x64, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x06, 0x68, 0x61, 0x6c, 0x62, 0x69, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x61, 0x6c, 0x66, 0x6c,
	0x6f, 0x61, 0x74, 0x18, 0x65, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x68, 0x61, 0x6c, 0x66, 0x6c,
	0x6f, 0x61, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x61, 0x6c, 0x73, 0x33, 0x32, 0x18, 0x66, 0x20,
	0x01, 0x28, 0x0f, 0x52, 0x06, 0x68, 0x61, 0x6c, 0x73, 0x33, 0x32, 0x12, 0x16, 0x0a, 0x06, 0x68,
	0x61, 0x6c, 0x75, 0x33, 0x32, 0x18, 0x67, 0x20, 0x01, 0x28, 0x07, 0x52, 0x06, 0x68, 0x61, 0x6c,
	0x75, 0x33, 0x32, 0x12, 0x17, 0x0a, 0x07, 0x76, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x78,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x76, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07,
	0x76, 0x5f, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x18, 0x82, 0x01, 0x20, 0x01, 0x28, 0x0f, 0x52, 0x06,
	0x76, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x5f, 0x69, 0x6e, 0x74, 0x36,
	0x34, 0x18, 0x8c, 0x01, 0x20, 0x01, 0x28, 0x10, 0x52, 0x06, 0x76, 0x49, 0x6e, 0x74, 0x36, 0x34,
	0x12, 0x1a, 0x0a, 0x08, 0x76, 0x5f, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x18, 0x96, 0x01, 0x20,
	0x01, 0x28, 0x07, 0x52, 0x07, 0x76, 0x55, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x12, 0x1a, 0x0a, 0x08,
	0x76, 0x5f, 0x75, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x18, 0xa0, 0x01, 0x20, 0x01, 0x28, 0x06, 0x52,
	0x07, 0x76, 0x55, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x12, 0x1a, 0x0a, 0x08, 0x76, 0x5f, 0x64, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x18, 0xaa, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x07, 0x76, 0x44, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x12, 0x21, 0x0a, 0x08, 0x76, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67,
	0x18, 0xb4, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x05, 0x92, 0x3f, 0x02, 0x08, 0x29, 0x52, 0x07,
	0x76, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x5f, 0x62, 0x6f, 0x6f,
	0x6c, 0x18, 0xbe, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x76, 0x42, 0x6f, 0x6f, 0x6c, 0x12,
	0x2f, 0x0a, 0x05, 0x63, 0x61, 0x72, 0x74, 0x65, 0x18, 0xc8, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2e, 0x50, 0x6d,
	0x43, 0x61, 0x72, 0x74, 0x65, 0x73, 0x69, 0x61, 0x6e, 0x52, 0x05, 0x63, 0x61, 0x72, 0x74, 0x65,
	0x12, 0x29, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x65, 0x18, 0xdc, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2e, 0x45, 0x6d,
	0x63, 0x50, 0x6f, 0x73, 0x65, 0x52, 0x04, 0x70, 0x6f, 0x73, 0x65, 0x3a, 0x06, 0x92, 0x3f, 0x03,
	0x48, 0xdc, 0x0b, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x6b, 0x69, 0x74, 0x2f, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2d, 0x67, 0x6f,
}

var (
	file_machinetalk_protobuf_value_proto_rawDescOnce sync.Once
	file_machinetalk_protobuf_value_proto_rawDescData = file_machinetalk_protobuf_value_proto_rawDesc
)

func file_machinetalk_protobuf_value_proto_rawDescGZIP() []byte {
	file_machinetalk_protobuf_value_proto_rawDescOnce.Do(func() {
		file_machinetalk_protobuf_value_proto_rawDescData = protoimpl.X.CompressGZIP(file_machinetalk_protobuf_value_proto_rawDescData)
	})
	return file_machinetalk_protobuf_value_proto_rawDescData
}

var file_machinetalk_protobuf_value_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_machinetalk_protobuf_value_proto_goTypes = []interface{}{
	(*Value)(nil),       // 0: machinetalk.Value
	(ValueType)(0),      // 1: machinetalk.ValueType
	(*PmCartesian)(nil), // 2: machinetalk.PmCartesian
	(*EmcPose)(nil),     // 3: machinetalk.EmcPose
}
var file_machinetalk_protobuf_value_proto_depIdxs = []int32{
	1, // 0: machinetalk.Value.type:type_name -> machinetalk.ValueType
	2, // 1: machinetalk.Value.carte:type_name -> machinetalk.PmCartesian
	3, // 2: machinetalk.Value.pose:type_name -> machinetalk.EmcPose
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_machinetalk_protobuf_value_proto_init() }
func file_machinetalk_protobuf_value_proto_init() {
	if File_machinetalk_protobuf_value_proto != nil {
		return
	}
	file_machinetalk_protobuf_nanopb_proto_init()
	file_machinetalk_protobuf_emcclass_proto_init()
	file_machinetalk_protobuf_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_machinetalk_protobuf_value_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Value); i {
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
			RawDescriptor: file_machinetalk_protobuf_value_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_machinetalk_protobuf_value_proto_goTypes,
		DependencyIndexes: file_machinetalk_protobuf_value_proto_depIdxs,
		MessageInfos:      file_machinetalk_protobuf_value_proto_msgTypes,
	}.Build()
	File_machinetalk_protobuf_value_proto = out.File
	file_machinetalk_protobuf_value_proto_rawDesc = nil
	file_machinetalk_protobuf_value_proto_goTypes = nil
	file_machinetalk_protobuf_value_proto_depIdxs = nil
}