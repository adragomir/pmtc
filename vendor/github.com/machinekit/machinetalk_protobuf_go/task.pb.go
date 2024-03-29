// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: machinetalk/protobuf/task.proto

// see README.msgid
// msgid base: 1200

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

type TaskPlanExecute struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command *string `protobuf:"bytes,10,opt,name=command" json:"command,omitempty"` // "MDI"
	Line    *int32  `protobuf:"fixed32,30,opt,name=line" json:"line,omitempty"`
}

func (x *TaskPlanExecute) Reset() {
	*x = TaskPlanExecute{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPlanExecute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPlanExecute) ProtoMessage() {}

func (x *TaskPlanExecute) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPlanExecute.ProtoReflect.Descriptor instead.
func (*TaskPlanExecute) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{0}
}

func (x *TaskPlanExecute) GetCommand() string {
	if x != nil && x.Command != nil {
		return *x.Command
	}
	return ""
}

func (x *TaskPlanExecute) GetLine() int32 {
	if x != nil && x.Line != nil {
		return *x.Line
	}
	return 0
}

type TaskPlanBlockDelete struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *bool `protobuf:"varint,10,req,name=state" json:"state,omitempty"`
}

func (x *TaskPlanBlockDelete) Reset() {
	*x = TaskPlanBlockDelete{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPlanBlockDelete) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPlanBlockDelete) ProtoMessage() {}

func (x *TaskPlanBlockDelete) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPlanBlockDelete.ProtoReflect.Descriptor instead.
func (*TaskPlanBlockDelete) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{1}
}

func (x *TaskPlanBlockDelete) GetState() bool {
	if x != nil && x.State != nil {
		return *x.State
	}
	return false
}

type TaskPlanOptionalStop struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *bool `protobuf:"varint,10,req,name=state" json:"state,omitempty"`
}

func (x *TaskPlanOptionalStop) Reset() {
	*x = TaskPlanOptionalStop{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPlanOptionalStop) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPlanOptionalStop) ProtoMessage() {}

func (x *TaskPlanOptionalStop) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPlanOptionalStop.ProtoReflect.Descriptor instead.
func (*TaskPlanOptionalStop) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{2}
}

func (x *TaskPlanOptionalStop) GetState() bool {
	if x != nil && x.State != nil {
		return *x.State
	}
	return false
}

type TaskPlanOpen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filename *string `protobuf:"bytes,10,req,name=filename" json:"filename,omitempty"`
}

func (x *TaskPlanOpen) Reset() {
	*x = TaskPlanOpen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPlanOpen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPlanOpen) ProtoMessage() {}

func (x *TaskPlanOpen) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPlanOpen.ProtoReflect.Descriptor instead.
func (*TaskPlanOpen) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{3}
}

func (x *TaskPlanOpen) GetFilename() string {
	if x != nil && x.Filename != nil {
		return *x.Filename
	}
	return ""
}

type TaskPlanReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cmd      *ContainerType `protobuf:"varint,10,req,name=cmd,enum=machinetalk.ContainerType" json:"cmd,omitempty"`
	Errormsg *string        `protobuf:"bytes,20,opt,name=errormsg" json:"errormsg,omitempty"`
}

func (x *TaskPlanReply) Reset() {
	*x = TaskPlanReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPlanReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPlanReply) ProtoMessage() {}

func (x *TaskPlanReply) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPlanReply.ProtoReflect.Descriptor instead.
func (*TaskPlanReply) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{4}
}

func (x *TaskPlanReply) GetCmd() ContainerType {
	if x != nil && x.Cmd != nil {
		return *x.Cmd
	}
	return ContainerType_MT_RTMESSAGE
}

func (x *TaskPlanReply) GetErrormsg() string {
	if x != nil && x.Errormsg != nil {
		return *x.Errormsg
	}
	return ""
}

// Ticket msgs
type TaskReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ticket *uint32 `protobuf:"fixed32,10,req,name=ticket" json:"ticket,omitempty"`
}

func (x *TaskReply) Reset() {
	*x = TaskReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskReply) ProtoMessage() {}

func (x *TaskReply) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskReply.ProtoReflect.Descriptor instead.
func (*TaskReply) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{5}
}

func (x *TaskReply) GetTicket() uint32 {
	if x != nil && x.Ticket != nil {
		return *x.Ticket
	}
	return 0
}

// signal completion of a particular ticket
type TicketUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cticket *uint32     `protobuf:"fixed32,10,req,name=cticket" json:"cticket,omitempty"`
	Status  *RCS_STATUS `protobuf:"varint,20,req,name=status,enum=machinetalk.RCS_STATUS" json:"status,omitempty"`
	Text    *string     `protobuf:"bytes,30,opt,name=text" json:"text,omitempty"`
}

func (x *TicketUpdate) Reset() {
	*x = TicketUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_machinetalk_protobuf_task_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TicketUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TicketUpdate) ProtoMessage() {}

func (x *TicketUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_machinetalk_protobuf_task_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TicketUpdate.ProtoReflect.Descriptor instead.
func (*TicketUpdate) Descriptor() ([]byte, []int) {
	return file_machinetalk_protobuf_task_proto_rawDescGZIP(), []int{6}
}

func (x *TicketUpdate) GetCticket() uint32 {
	if x != nil && x.Cticket != nil {
		return *x.Cticket
	}
	return 0
}

func (x *TicketUpdate) GetStatus() RCS_STATUS {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return RCS_STATUS_UNINITIALIZED_STATUS
}

func (x *TicketUpdate) GetText() string {
	if x != nil && x.Text != nil {
		return *x.Text
	}
	return ""
}

var File_machinetalk_protobuf_task_proto protoreflect.FileDescriptor

var file_machinetalk_protobuf_task_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0b, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x1a, 0x20,
	0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x21, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x61, 0x6e, 0x6f, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x47, 0x0a, 0x0f, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x6c, 0x61, 0x6e, 0x45,
	0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x0f, 0x52, 0x04,
	0x6c, 0x69, 0x6e, 0x65, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xb0, 0x09, 0x22, 0x33, 0x0a, 0x13,
	0x54, 0x61, 0x73, 0x6b, 0x50, 0x6c, 0x61, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x0a, 0x20, 0x02,
	0x28, 0x08, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xb1,
	0x09, 0x22, 0x34, 0x0a, 0x14, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x6c, 0x61, 0x6e, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x53, 0x74, 0x6f, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x08, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x3a,
	0x06, 0x92, 0x3f, 0x03, 0x48, 0xb2, 0x09, 0x22, 0x32, 0x0a, 0x0c, 0x54, 0x61, 0x73, 0x6b, 0x50,
	0x6c, 0x61, 0x6e, 0x4f, 0x70, 0x65, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e,
	0x61, 0x6d, 0x65, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xb3, 0x09, 0x22, 0x61, 0x0a, 0x0d, 0x54,
	0x61, 0x73, 0x6b, 0x50, 0x6c, 0x61, 0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x03,
	0x63, 0x6d, 0x64, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x6d, 0x61, 0x63, 0x68,
	0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65,
	0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x6d, 0x73, 0x67, 0x18, 0x14, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x6d, 0x73, 0x67, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xb4, 0x09, 0x22, 0x2b,
	0x0a, 0x09, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x74,
	0x69, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x07, 0x52, 0x06, 0x74, 0x69, 0x63,
	0x6b, 0x65, 0x74, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48, 0xb5, 0x09, 0x22, 0x75, 0x0a, 0x0c, 0x54,
	0x69, 0x63, 0x6b, 0x65, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x74, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x07, 0x52, 0x07, 0x63, 0x74,
	0x69, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x2f, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x14, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x74,
	0x61, 0x6c, 0x6b, 0x2e, 0x52, 0x43, 0x53, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x1e,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x3a, 0x06, 0x92, 0x3f, 0x03, 0x48,
	0xb6, 0x09, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x6b, 0x69, 0x74, 0x2f, 0x6d, 0x61, 0x63, 0x68,
	0x69, 0x6e, 0x65, 0x74, 0x61, 0x6c, 0x6b, 0x2d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2d, 0x67, 0x6f,
}

var (
	file_machinetalk_protobuf_task_proto_rawDescOnce sync.Once
	file_machinetalk_protobuf_task_proto_rawDescData = file_machinetalk_protobuf_task_proto_rawDesc
)

func file_machinetalk_protobuf_task_proto_rawDescGZIP() []byte {
	file_machinetalk_protobuf_task_proto_rawDescOnce.Do(func() {
		file_machinetalk_protobuf_task_proto_rawDescData = protoimpl.X.CompressGZIP(file_machinetalk_protobuf_task_proto_rawDescData)
	})
	return file_machinetalk_protobuf_task_proto_rawDescData
}

var file_machinetalk_protobuf_task_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_machinetalk_protobuf_task_proto_goTypes = []interface{}{
	(*TaskPlanExecute)(nil),      // 0: machinetalk.TaskPlanExecute
	(*TaskPlanBlockDelete)(nil),  // 1: machinetalk.TaskPlanBlockDelete
	(*TaskPlanOptionalStop)(nil), // 2: machinetalk.TaskPlanOptionalStop
	(*TaskPlanOpen)(nil),         // 3: machinetalk.TaskPlanOpen
	(*TaskPlanReply)(nil),        // 4: machinetalk.TaskPlanReply
	(*TaskReply)(nil),            // 5: machinetalk.TaskReply
	(*TicketUpdate)(nil),         // 6: machinetalk.TicketUpdate
	(ContainerType)(0),           // 7: machinetalk.ContainerType
	(RCS_STATUS)(0),              // 8: machinetalk.RCS_STATUS
}
var file_machinetalk_protobuf_task_proto_depIdxs = []int32{
	7, // 0: machinetalk.TaskPlanReply.cmd:type_name -> machinetalk.ContainerType
	8, // 1: machinetalk.TicketUpdate.status:type_name -> machinetalk.RCS_STATUS
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_machinetalk_protobuf_task_proto_init() }
func file_machinetalk_protobuf_task_proto_init() {
	if File_machinetalk_protobuf_task_proto != nil {
		return
	}
	file_machinetalk_protobuf_types_proto_init()
	file_machinetalk_protobuf_nanopb_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_machinetalk_protobuf_task_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPlanExecute); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPlanBlockDelete); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPlanOptionalStop); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPlanOpen); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPlanReply); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskReply); i {
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
		file_machinetalk_protobuf_task_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TicketUpdate); i {
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
			RawDescriptor: file_machinetalk_protobuf_task_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_machinetalk_protobuf_task_proto_goTypes,
		DependencyIndexes: file_machinetalk_protobuf_task_proto_depIdxs,
		MessageInfos:      file_machinetalk_protobuf_task_proto_msgTypes,
	}.Build()
	File_machinetalk_protobuf_task_proto = out.File
	file_machinetalk_protobuf_task_proto_rawDesc = nil
	file_machinetalk_protobuf_task_proto_goTypes = nil
	file_machinetalk_protobuf_task_proto_depIdxs = nil
}
