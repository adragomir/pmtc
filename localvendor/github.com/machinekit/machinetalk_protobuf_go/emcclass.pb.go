// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: machinetalk/protobuf/emcclass.proto

// see README.msgid
// msgid base: 300

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

type PmCartesian struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	X *float64 `protobuf:"fixed64,10,opt,name=x" json:"x,omitempty"`
	Y *float64 `protobuf:"fixed64,20,opt,name=y" json:"y,omitempty"`
	Z *float64 `protobuf:"fixed64,30,opt,name=z" json:"z,omitempty"`
}

func (x *PmCartesian) Reset() {
	*x = PmCartesian{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_emcclass_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PmCartesian) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PmCartesian) ProtoMessage() {}

func (x *PmCartesian) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_emcclass_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PmCartesian.ProtoReflect.Descriptor instead.
func (*PmCartesian) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_emcclass_proto_rawDescGZIP(), []int{0}
}

func (x *PmCartesian) GetX() float64 {
	if x != nil && x.X != nil {
		return *x.X
	}
	return 0
}

func (x *PmCartesian) GetY() float64 {
	if x != nil && x.Y != nil {
		return *x.Y
	}
	return 0
}

func (x *PmCartesian) GetZ() float64 {
	if x != nil && x.Z != nil {
		return *x.Z
	}
	return 0
}

type EmcPose struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tran *PmCartesian `protobuf:"bytes,10,req,name=tran" json:"tran,omitempty"`
	A    *float64     `protobuf:"fixed64,20,opt,name=a" json:"a,omitempty"`
	B    *float64     `protobuf:"fixed64,30,opt,name=b" json:"b,omitempty"`
	C    *float64     `protobuf:"fixed64,40,opt,name=c" json:"c,omitempty"`
	U    *float64     `protobuf:"fixed64,50,opt,name=u" json:"u,omitempty"`
	V    *float64     `protobuf:"fixed64,60,opt,name=v" json:"v,omitempty"`
	W    *float64     `protobuf:"fixed64,70,opt,name=w" json:"w,omitempty"`
}

func (x *EmcPose) Reset() {
	*x = EmcPose{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_emcclass_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmcPose) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmcPose) ProtoMessage() {}

func (x *EmcPose) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_emcclass_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmcPose.ProtoReflect.Descriptor instead.
func (*EmcPose) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_emcclass_proto_rawDescGZIP(), []int{1}
}

func (x *EmcPose) GetTran() *PmCartesian {
	if x != nil {
		return x.Tran
	}
	return nil
}

func (x *EmcPose) GetA() float64 {
	if x != nil && x.A != nil {
		return *x.A
	}
	return 0
}

func (x *EmcPose) GetB() float64 {
	if x != nil && x.B != nil {
		return *x.B
	}
	return 0
}

func (x *EmcPose) GetC() float64 {
	if x != nil && x.C != nil {
		return *x.C
	}
	return 0
}

func (x *EmcPose) GetU() float64 {
	if x != nil && x.U != nil {
		return *x.U
	}
	return 0
}

func (x *EmcPose) GetV() float64 {
	if x != nil && x.V != nil {
		return *x.V
	}
	return 0
}

func (x *EmcPose) GetW() float64 {
	if x != nil && x.W != nil {
		return *x.W
	}
	return 0
}

var File_machinetalk_protobuf_emcclass_proto protoreflect.FileDescriptor

var file_machinetalk_protobuf_emcclass_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x63, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61,
	0x6c, 0x6b, 0x1a, 0x21, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x61, 0x6e, 0x6f, 0x70, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3f, 0x0a, 0x0b, 0x50, 0x6d, 0x43, 0x61, 0x72, 0x74, 0x65,
	0x73, 0x69, 0x61, 0x6e, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x14, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x79,
	0x12, 0x0c, 0x0a, 0x01, 0x7a, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x7a, 0x3a, 0x06,
	0x92, 0x3f, 0x03, 0x48, 0xac, 0x02, 0x22, 0x93, 0x01, 0x0a, 0x07, 0x45, 0x6d, 0x63, 0x50, 0x6f,
	0x73, 0x65, 0x12, 0x2c, 0x0a, 0x04, 0x74, 0x72, 0x61, 0x6e, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2e, 0x50,
	0x6d, 0x43, 0x61, 0x72, 0x74, 0x65, 0x73, 0x69, 0x61, 0x6e, 0x52, 0x04, 0x74, 0x72, 0x61, 0x6e,
	0x12, 0x0c, 0x0a, 0x01, 0x61, 0x18, 0x14, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x61, 0x12, 0x0c,
	0x0a, 0x01, 0x62, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x62, 0x12, 0x0c, 0x0a, 0x01,
	0x63, 0x18, 0x28, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x63, 0x12, 0x0c, 0x0a, 0x01, 0x75, 0x18,
	0x32, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x75, 0x12, 0x0c, 0x0a, 0x01, 0x76, 0x18, 0x3c, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x01, 0x76, 0x12, 0x0c, 0x0a, 0x01, 0x77, 0x18, 0x46, 0x20, 0x01, 0x28,
	0x01, 0x52, 0x01, 0x77, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xad, 0x02, 0x42, 0x2f, 0x5a, 0x2d,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x6b, 0x69, 0x74, 0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c,
	0x6b, 0x2d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2d, 0x67, 0x6f,
}

var (
	file_machinetalk_protobuf_emcclass_proto_rawDescOnce sync.Once
	file_machinetalk_protobuf_emcclass_proto_rawDescData = file_machinetalk_protobuf_emcclass_proto_rawDesc
)

func file_machinetalk_protobuf_emcclass_proto_rawDescGZIP() []byte {
	file_machinetalk_protobuf_emcclass_proto_rawDescOnce.Do(func() {
		file_machinetalk_protobuf_emcclass_proto_rawDescData = protoimpl.X.CompressGZIP(file_machinetalk_protobuf_emcclass_proto_rawDescData)
	})
	return file_machinetalk_protobuf_emcclass_proto_rawDescData
}

var file_machinetalk_protobuf_emcclass_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_machinetalk_protobuf_emcclass_proto_goTypes = []interface{}{
	(*PmCartesian)(nil), // 0: machinetalk.PmCartesian
	(*EmcPose)(nil),     // 1: machinetalk.EmcPose
}
var file_machinetalk_protobuf_emcclass_proto_depIdxs = []int32{
	0, // 0: machinetalk.EmcPose.tran:type_name -> machinetalk.PmCartesian
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_machinetalk_protobuf_emcclass_proto_init() }
func file_machinetalk_protobuf_emcclass_proto_init() {
	if File_machinetalk_protobuf_emcclass_proto != nil {
		return
	}
	file_machinetalk_protobuf_nanopb_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_machinetalk_protobuf_emcclass_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PmCartesian); i {
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
		file_machinetalk_protobuf_emcclass_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmcPose); i {
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
			RawDescriptor: file_machinetalk_protobuf_emcclass_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_machinetalk_protobuf_emcclass_proto_goTypes,
		DependencyIndexes: file_machinetalk_protobuf_emcclass_proto_depIdxs,
		MessageInfos:      file_machinetalk_protobuf_emcclass_proto_msgTypes,
	}.Build()
	File_machinetalk_protobuf_emcclass_proto = out.File
	file_machinetalk_protobuf_emcclass_proto_rawDesc = nil
	file_machinetalk_protobuf_emcclass_proto_goTypes = nil
	file_machinetalk_protobuf_emcclass_proto_depIdxs = nil
}
