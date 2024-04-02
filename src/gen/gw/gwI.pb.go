// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: dsl/gwI.proto

package gw

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

type GidState int32

const (
	// 连接
	GidState_Conn GidState = 0
	// 拉取数据
	GidState_GLast GidState = 1
	// 断开
	GidState_Disc GidState = 2
)

// Enum value maps for GidState.
var (
	GidState_name = map[int32]string{
		0: "Conn",
		1: "GLast",
		2: "Disc",
	}
	GidState_value = map[string]int32{
		"Conn":  0,
		"GLast": 1,
		"Disc":  2,
	}
)

func (x GidState) Enum() *GidState {
	p := new(GidState)
	*p = x
	return p
}

func (x GidState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GidState) Descriptor() protoreflect.EnumDescriptor {
	return file_dsl_gwI_proto_enumTypes[0].Descriptor()
}

func (GidState) Type() protoreflect.EnumType {
	return &file_dsl_gwI_proto_enumTypes[0]
}

func (x GidState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GidState.Descriptor instead.
func (GidState) EnumDescriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{0}
}

type CidGidReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cid    int64    `protobuf:"varint,1,opt,name=cid,proto3" json:"cid,omitempty"`
	Gid    string   `protobuf:"bytes,2,opt,name=gid,proto3" json:"gid,omitempty"`
	Unique string   `protobuf:"bytes,3,opt,name=unique,proto3" json:"unique,omitempty"`
	State  GidState `protobuf:"varint,4,opt,name=state,proto3,enum=gw.GidState" json:"state,omitempty"`
}

func (x *CidGidReq) Reset() {
	*x = CidGidReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CidGidReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CidGidReq) ProtoMessage() {}

func (x *CidGidReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CidGidReq.ProtoReflect.Descriptor instead.
func (*CidGidReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{0}
}

func (x *CidGidReq) GetCid() int64 {
	if x != nil {
		return x.Cid
	}
	return 0
}

func (x *CidGidReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *CidGidReq) GetUnique() string {
	if x != nil {
		return x.Unique
	}
	return ""
}

func (x *CidGidReq) GetState() GidState {
	if x != nil {
		return x.State
	}
	return GidState_Conn
}

type RepReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cid int64 `protobuf:"varint,1,opt,name=cid,proto3" json:"cid,omitempty"`
	// 写入类型
	Req int32 `protobuf:"varint,2,opt,name=req,proto3" json:"req,omitempty"`
	// 写入连接
	Uri string `protobuf:"bytes,3,opt,name=uri,proto3" json:"uri,omitempty"`
	// 写入链接压缩
	UriI int32 `protobuf:"varint,4,opt,name=uriI,proto3" json:"uriI,omitempty"`
	// 写入数据
	Data []byte `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
	// 已压缩
	CDid int32 `protobuf:"varint,6,opt,name=cDid,proto3" json:"cDid,omitempty"` // 1 已压缩 2 已尝试压缩，未能压缩
	// 加密
	Encry bool `protobuf:"varint,7,opt,name=encry,proto3" json:"encry,omitempty"`
	// 独立数据
	Isolate bool `protobuf:"varint,8,opt,name=Isolate,proto3" json:"Isolate,omitempty"`
}

func (x *RepReq) Reset() {
	*x = RepReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepReq) ProtoMessage() {}

func (x *RepReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepReq.ProtoReflect.Descriptor instead.
func (*RepReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{1}
}

func (x *RepReq) GetCid() int64 {
	if x != nil {
		return x.Cid
	}
	return 0
}

func (x *RepReq) GetReq() int32 {
	if x != nil {
		return x.Req
	}
	return 0
}

func (x *RepReq) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *RepReq) GetUriI() int32 {
	if x != nil {
		return x.UriI
	}
	return 0
}

func (x *RepReq) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *RepReq) GetCDid() int32 {
	if x != nil {
		return x.CDid
	}
	return 0
}

func (x *RepReq) GetEncry() bool {
	if x != nil {
		return x.Encry
	}
	return false
}

func (x *RepReq) GetIsolate() bool {
	if x != nil {
		return x.Isolate
	}
	return false
}

type Msg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uri  string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Msg) Reset() {
	*x = Msg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Msg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Msg) ProtoMessage() {}

func (x *Msg) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Msg.ProtoReflect.Descriptor instead.
func (*Msg) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{2}
}

func (x *Msg) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *Msg) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ILastReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cid        int64  `protobuf:"varint,1,opt,name=cid,proto3" json:"cid,omitempty"`
	Gid        string `protobuf:"bytes,2,opt,name=gid,proto3" json:"gid,omitempty"`
	ConnVer    int32  `protobuf:"varint,3,opt,name=connVer,proto3" json:"connVer,omitempty"`
	Continuous bool   `protobuf:"varint,4,opt,name=continuous,proto3" json:"continuous,omitempty"`
}

func (x *ILastReq) Reset() {
	*x = ILastReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ILastReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ILastReq) ProtoMessage() {}

func (x *ILastReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ILastReq.ProtoReflect.Descriptor instead.
func (*ILastReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{3}
}

func (x *ILastReq) GetCid() int64 {
	if x != nil {
		return x.Cid
	}
	return 0
}

func (x *ILastReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *ILastReq) GetConnVer() int32 {
	if x != nil {
		return x.ConnVer
	}
	return 0
}

func (x *ILastReq) GetContinuous() bool {
	if x != nil {
		return x.Continuous
	}
	return false
}

type IGQueueReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid    string `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Cid    int64  `protobuf:"varint,2,opt,name=cid,proto3" json:"cid,omitempty"`
	Unique string `protobuf:"bytes,3,opt,name=unique,proto3" json:"unique,omitempty"`
	Clear  bool   `protobuf:"varint,4,opt,name=clear,proto3" json:"clear,omitempty"`
}

func (x *IGQueueReq) Reset() {
	*x = IGQueueReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IGQueueReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IGQueueReq) ProtoMessage() {}

func (x *IGQueueReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IGQueueReq.ProtoReflect.Descriptor instead.
func (*IGQueueReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{4}
}

func (x *IGQueueReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *IGQueueReq) GetCid() int64 {
	if x != nil {
		return x.Cid
	}
	return 0
}

func (x *IGQueueReq) GetUnique() string {
	if x != nil {
		return x.Unique
	}
	return ""
}

func (x *IGQueueReq) GetClear() bool {
	if x != nil {
		return x.Clear
	}
	return false
}

type IGClearReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid   string `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Queue bool   `protobuf:"varint,2,opt,name=queue,proto3" json:"queue,omitempty"`
	Last  bool   `protobuf:"varint,3,opt,name=last,proto3" json:"last,omitempty"`
}

func (x *IGClearReq) Reset() {
	*x = IGClearReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IGClearReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IGClearReq) ProtoMessage() {}

func (x *IGClearReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IGClearReq.ProtoReflect.Descriptor instead.
func (*IGClearReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{5}
}

func (x *IGClearReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *IGClearReq) GetQueue() bool {
	if x != nil {
		return x.Queue
	}
	return false
}

func (x *IGClearReq) GetLast() bool {
	if x != nil {
		return x.Last
	}
	return false
}

type IGPushAReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid  string `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Id   int64  `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	Succ bool   `protobuf:"varint,3,opt,name=succ,proto3" json:"succ,omitempty"`
}

func (x *IGPushAReq) Reset() {
	*x = IGPushAReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IGPushAReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IGPushAReq) ProtoMessage() {}

func (x *IGPushAReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IGPushAReq.ProtoReflect.Descriptor instead.
func (*IGPushAReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{6}
}

func (x *IGPushAReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *IGPushAReq) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *IGPushAReq) GetSucc() bool {
	if x != nil {
		return x.Succ
	}
	return false
}

type ReadReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid    string `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Tid    string `protobuf:"bytes,2,opt,name=tid,proto3" json:"tid,omitempty"`
	LastId int64  `protobuf:"varint,3,opt,name=lastId,proto3" json:"lastId,omitempty"`
}

func (x *ReadReq) Reset() {
	*x = ReadReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadReq) ProtoMessage() {}

func (x *ReadReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadReq.ProtoReflect.Descriptor instead.
func (*ReadReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{7}
}

func (x *ReadReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *ReadReq) GetTid() string {
	if x != nil {
		return x.Tid
	}
	return ""
}

func (x *ReadReq) GetLastId() int64 {
	if x != nil {
		return x.LastId
	}
	return 0
}

type UnreadReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid    string `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Tid    string `protobuf:"bytes,2,opt,name=tid,proto3" json:"tid,omitempty"`
	Num    int32  `protobuf:"varint,3,opt,name=num,proto3" json:"num,omitempty"`
	LastId int64  `protobuf:"varint,4,opt,name=lastId,proto3" json:"lastId,omitempty"`
	Uri    string `protobuf:"bytes,5,opt,name=uri,proto3" json:"uri,omitempty"`
	Data   []byte `protobuf:"bytes,6,opt,name=data,proto3" json:"data,omitempty"`
	Entry  bool   `protobuf:"varint,7,opt,name=entry,proto3" json:"entry,omitempty"`
}

func (x *UnreadReq) Reset() {
	*x = UnreadReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnreadReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnreadReq) ProtoMessage() {}

func (x *UnreadReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnreadReq.ProtoReflect.Descriptor instead.
func (*UnreadReq) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{8}
}

func (x *UnreadReq) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *UnreadReq) GetTid() string {
	if x != nil {
		return x.Tid
	}
	return ""
}

func (x *UnreadReq) GetNum() int32 {
	if x != nil {
		return x.Num
	}
	return 0
}

func (x *UnreadReq) GetLastId() int64 {
	if x != nil {
		return x.LastId
	}
	return 0
}

func (x *UnreadReq) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *UnreadReq) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *UnreadReq) GetEntry() bool {
	if x != nil {
		return x.Entry
	}
	return false
}

type UnreadReqs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Gid  string       `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Reqs []*UnreadReq `protobuf:"bytes,2,rep,name=reqs,proto3" json:"reqs,omitempty"`
}

func (x *UnreadReqs) Reset() {
	*x = UnreadReqs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsl_gwI_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnreadReqs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnreadReqs) ProtoMessage() {}

func (x *UnreadReqs) ProtoReflect() protoreflect.Message {
	mi := &file_dsl_gwI_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnreadReqs.ProtoReflect.Descriptor instead.
func (*UnreadReqs) Descriptor() ([]byte, []int) {
	return file_dsl_gwI_proto_rawDescGZIP(), []int{9}
}

func (x *UnreadReqs) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *UnreadReqs) GetReqs() []*UnreadReq {
	if x != nil {
		return x.Reqs
	}
	return nil
}

var File_dsl_gwI_proto protoreflect.FileDescriptor

var file_dsl_gwI_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x64, 0x73, 0x6c, 0x2f, 0x67, 0x77, 0x49, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x67, 0x77, 0x1a, 0x0c, 0x64, 0x73, 0x6c, 0x2f, 0x67, 0x77, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x6b, 0x0a, 0x09, 0x43, 0x69, 0x64, 0x47, 0x69, 0x64, 0x52, 0x65, 0x71, 0x12, 0x10,
	0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x63, 0x69, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67,
	0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x75, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x12, 0x22, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x47,
	0x69, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0xaa,
	0x01, 0x0a, 0x06, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x63, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x72,
	0x65, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x72, 0x65, 0x71, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x69, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12,
	0x12, 0x0a, 0x04, 0x75, 0x72, 0x69, 0x49, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x75,
	0x72, 0x69, 0x49, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x44, 0x69, 0x64, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x44, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x6e, 0x63, 0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x65, 0x6e, 0x63, 0x72,
	0x79, 0x12, 0x18, 0x0a, 0x07, 0x49, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x49, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x22, 0x2b, 0x0a, 0x03, 0x4d,
	0x73, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x69, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x68, 0x0a, 0x08, 0x49, 0x4c, 0x61, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x03, 0x63, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x6e,
	0x56, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x6e, 0x56,
	0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x69, 0x6e, 0x75, 0x6f, 0x75, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x69, 0x6e, 0x75, 0x6f,
	0x75, 0x73, 0x22, 0x5e, 0x0a, 0x0a, 0x49, 0x47, 0x51, 0x75, 0x65, 0x75, 0x65, 0x52, 0x65, 0x71,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67,
	0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x03, 0x63, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6c, 0x65, 0x61, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x63, 0x6c, 0x65,
	0x61, 0x72, 0x22, 0x48, 0x0a, 0x0a, 0x49, 0x47, 0x43, 0x6c, 0x65, 0x61, 0x72, 0x52, 0x65, 0x71,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67,
	0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x05, 0x71, 0x75, 0x65, 0x75, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x61, 0x73, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x42, 0x0a, 0x0a,
	0x49, 0x47, 0x50, 0x75, 0x73, 0x68, 0x41, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x75, 0x63, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x73, 0x75, 0x63, 0x63,
	0x22, 0x45, 0x0a, 0x07, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x67,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x74, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x69, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x6c, 0x61, 0x73, 0x74, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x06, 0x6c, 0x61, 0x73, 0x74, 0x49, 0x64, 0x22, 0x95, 0x01, 0x0a, 0x09, 0x55, 0x6e, 0x72, 0x65,
	0x61, 0x64, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6e, 0x75, 0x6d,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6e, 0x75, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x6c,
	0x61, 0x73, 0x74, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6c, 0x61, 0x73,
	0x74, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6e, 0x74,
	0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x22,
	0x41, 0x0a, 0x0a, 0x55, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x73, 0x12, 0x10, 0x0a,
	0x03, 0x67, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64, 0x12,
	0x21, 0x0a, 0x04, 0x72, 0x65, 0x71, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x67, 0x77, 0x2e, 0x55, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x52, 0x04, 0x72, 0x65,
	0x71, 0x73, 0x2a, 0x29, 0x0a, 0x08, 0x47, 0x69, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x08,
	0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x6e, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x4c, 0x61, 0x73,
	0x74, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x69, 0x73, 0x63, 0x10, 0x02, 0x32, 0xe4, 0x08,
	0x0a, 0x08, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x49, 0x12, 0x1d, 0x0a, 0x03, 0x75, 0x69,
	0x64, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x43, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0a, 0x2e,
	0x67, 0x77, 0x2e, 0x55, 0x49, 0x64, 0x52, 0x65, 0x70, 0x12, 0x21, 0x0a, 0x06, 0x6f, 0x6e, 0x6c,
	0x69, 0x6e, 0x65, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a,
	0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x52, 0x65, 0x70, 0x12, 0x24, 0x0a, 0x07,
	0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x12, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x69, 0x64,
	0x73, 0x52, 0x65, 0x71, 0x1a, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x73, 0x52,
	0x65, 0x70, 0x12, 0x22, 0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x0c, 0x2e, 0x67, 0x77,
	0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49,
	0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x04, 0x6b, 0x69, 0x63, 0x6b, 0x12, 0x0b,
	0x2e, 0x67, 0x77, 0x2e, 0x4b, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77,
	0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x05, 0x61, 0x6c, 0x69, 0x76,
	0x65, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x43, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e,
	0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x1e, 0x0a, 0x03, 0x72, 0x69,
	0x64, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x52, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e,
	0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x04, 0x72, 0x69,
	0x64, 0x73, 0x12, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x52, 0x69, 0x64, 0x73, 0x52, 0x65, 0x71, 0x1a,
	0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x1f, 0x0a, 0x04,
	0x63, 0x69, 0x64, 0x73, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x69, 0x64, 0x52, 0x65, 0x71,
	0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x43, 0x69, 0x64, 0x73, 0x52, 0x65, 0x70, 0x12, 0x24, 0x0a,
	0x06, 0x63, 0x69, 0x64, 0x47, 0x69, 0x64, 0x12, 0x0d, 0x2e, 0x67, 0x77, 0x2e, 0x43, 0x69, 0x64,
	0x47, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32,
	0x52, 0x65, 0x70, 0x12, 0x24, 0x0a, 0x06, 0x67, 0x69, 0x64, 0x43, 0x69, 0x64, 0x12, 0x0d, 0x2e,
	0x67, 0x77, 0x2e, 0x43, 0x69, 0x64, 0x47, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67,
	0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x21, 0x0a, 0x04, 0x63, 0x6f, 0x6e,
	0x6e, 0x12, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x43, 0x6f, 0x6e, 0x6e, 0x52, 0x65, 0x71, 0x1a,
	0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x21, 0x0a, 0x04,
	0x64, 0x69, 0x73, 0x63, 0x12, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x44, 0x69, 0x73, 0x63, 0x52,
	0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12,
	0x1e, 0x0a, 0x03, 0x72, 0x65, 0x70, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x52, 0x65, 0x70, 0x52,
	0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12,
	0x21, 0x0a, 0x04, 0x6c, 0x61, 0x73, 0x74, 0x12, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x4c, 0x61,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52,
	0x65, 0x70, 0x12, 0x20, 0x0a, 0x04, 0x70, 0x75, 0x73, 0x68, 0x12, 0x0b, 0x2e, 0x67, 0x77, 0x2e,
	0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33,
	0x32, 0x52, 0x65, 0x70, 0x12, 0x25, 0x0a, 0x06, 0x67, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x0e,
	0x2e, 0x67, 0x77, 0x2e, 0x49, 0x47, 0x51, 0x75, 0x65, 0x75, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x0b,
	0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x25, 0x0a, 0x06, 0x67,
	0x43, 0x6c, 0x65, 0x61, 0x72, 0x12, 0x0e, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x47, 0x43, 0x6c, 0x65,
	0x61, 0x72, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52,
	0x65, 0x70, 0x12, 0x24, 0x0a, 0x06, 0x67, 0x4c, 0x61, 0x73, 0x74, 0x73, 0x12, 0x0d, 0x2e, 0x67,
	0x77, 0x2e, 0x47, 0x4c, 0x61, 0x73, 0x74, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77,
	0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x05, 0x67, 0x4c, 0x61, 0x73,
	0x74, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e,
	0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x22, 0x0a, 0x05, 0x67, 0x50,
	0x75, 0x73, 0x68, 0x12, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65,
	0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x36, 0x34, 0x52, 0x65, 0x70, 0x12, 0x25,
	0x0a, 0x06, 0x67, 0x50, 0x75, 0x73, 0x68, 0x41, 0x12, 0x0e, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x47,
	0x50, 0x75, 0x73, 0x68, 0x41, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64,
	0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x04, 0x73, 0x65, 0x6e, 0x64, 0x12, 0x0b, 0x2e,
	0x67, 0x77, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e,
	0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x22, 0x0a, 0x05, 0x74, 0x50, 0x75, 0x73, 0x68,
	0x12, 0x0c, 0x2e, 0x67, 0x77, 0x2e, 0x54, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x71, 0x1a, 0x0b,
	0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x21, 0x0a, 0x06, 0x74,
	0x44, 0x69, 0x72, 0x74, 0x79, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x47, 0x69, 0x64, 0x52, 0x65,
	0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x22,
	0x0a, 0x07, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x73, 0x12, 0x0a, 0x2e, 0x67, 0x77, 0x2e, 0x47,
	0x69, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52,
	0x65, 0x70, 0x12, 0x25, 0x0a, 0x08, 0x73, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x64, 0x73, 0x12, 0x0c,
	0x2e, 0x67, 0x77, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x73, 0x52, 0x65, 0x70, 0x1a, 0x0b, 0x2e, 0x67,
	0x77, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x52, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x04, 0x72, 0x65, 0x61,
	0x64, 0x12, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b,
	0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x24, 0x0a, 0x06, 0x75,
	0x6e, 0x72, 0x65, 0x61, 0x64, 0x12, 0x0d, 0x2e, 0x67, 0x77, 0x2e, 0x55, 0x6e, 0x72, 0x65, 0x61,
	0x64, 0x52, 0x65, 0x71, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65,
	0x70, 0x12, 0x26, 0x0a, 0x07, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x73, 0x12, 0x0e, 0x2e, 0x67,
	0x77, 0x2e, 0x55, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x73, 0x1a, 0x0b, 0x2e, 0x67,
	0x77, 0x2e, 0x49, 0x64, 0x33, 0x32, 0x52, 0x65, 0x70, 0x12, 0x29, 0x0a, 0x0a, 0x75, 0x6e, 0x72,
	0x65, 0x61, 0x64, 0x54, 0x69, 0x64, 0x73, 0x12, 0x0e, 0x2e, 0x67, 0x77, 0x2e, 0x55, 0x6e, 0x72,
	0x65, 0x61, 0x64, 0x54, 0x69, 0x64, 0x73, 0x1a, 0x0b, 0x2e, 0x67, 0x77, 0x2e, 0x49, 0x64, 0x33,
	0x32, 0x52, 0x65, 0x70, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x67, 0x77, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dsl_gwI_proto_rawDescOnce sync.Once
	file_dsl_gwI_proto_rawDescData = file_dsl_gwI_proto_rawDesc
)

func file_dsl_gwI_proto_rawDescGZIP() []byte {
	file_dsl_gwI_proto_rawDescOnce.Do(func() {
		file_dsl_gwI_proto_rawDescData = protoimpl.X.CompressGZIP(file_dsl_gwI_proto_rawDescData)
	})
	return file_dsl_gwI_proto_rawDescData
}

var file_dsl_gwI_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_dsl_gwI_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_dsl_gwI_proto_goTypes = []interface{}{
	(GidState)(0),      // 0: gw.GidState
	(*CidGidReq)(nil),  // 1: gw.CidGidReq
	(*RepReq)(nil),     // 2: gw.RepReq
	(*Msg)(nil),        // 3: gw.Msg
	(*ILastReq)(nil),   // 4: gw.ILastReq
	(*IGQueueReq)(nil), // 5: gw.IGQueueReq
	(*IGClearReq)(nil), // 6: gw.IGClearReq
	(*IGPushAReq)(nil), // 7: gw.IGPushAReq
	(*ReadReq)(nil),    // 8: gw.ReadReq
	(*UnreadReq)(nil),  // 9: gw.UnreadReq
	(*UnreadReqs)(nil), // 10: gw.UnreadReqs
	(*CidReq)(nil),     // 11: gw.CidReq
	(*GidReq)(nil),     // 12: gw.GidReq
	(*GidsReq)(nil),    // 13: gw.GidsReq
	(*CloseReq)(nil),   // 14: gw.CloseReq
	(*KickReq)(nil),    // 15: gw.KickReq
	(*RidReq)(nil),     // 16: gw.RidReq
	(*RidsReq)(nil),    // 17: gw.RidsReq
	(*GConnReq)(nil),   // 18: gw.GConnReq
	(*GDiscReq)(nil),   // 19: gw.GDiscReq
	(*PushReq)(nil),    // 20: gw.PushReq
	(*GLastsReq)(nil),  // 21: gw.GLastsReq
	(*GPushReq)(nil),   // 22: gw.GPushReq
	(*SendReq)(nil),    // 23: gw.SendReq
	(*TPushReq)(nil),   // 24: gw.TPushReq
	(*ProdsRep)(nil),   // 25: gw.ProdsRep
	(*UnreadTids)(nil), // 26: gw.UnreadTids
	(*UIdRep)(nil),     // 27: gw.UIdRep
	(*BoolRep)(nil),    // 28: gw.BoolRep
	(*BoolsRep)(nil),   // 29: gw.BoolsRep
	(*Id32Rep)(nil),    // 30: gw.Id32Rep
	(*CidsRep)(nil),    // 31: gw.CidsRep
	(*Id64Rep)(nil),    // 32: gw.Id64Rep
}
var file_dsl_gwI_proto_depIdxs = []int32{
	0,  // 0: gw.CidGidReq.state:type_name -> gw.GidState
	9,  // 1: gw.UnreadReqs.reqs:type_name -> gw.UnreadReq
	11, // 2: gw.GatewayI.uid:input_type -> gw.CidReq
	12, // 3: gw.GatewayI.online:input_type -> gw.GidReq
	13, // 4: gw.GatewayI.onlines:input_type -> gw.GidsReq
	14, // 5: gw.GatewayI.close:input_type -> gw.CloseReq
	15, // 6: gw.GatewayI.kick:input_type -> gw.KickReq
	11, // 7: gw.GatewayI.alive:input_type -> gw.CidReq
	16, // 8: gw.GatewayI.rid:input_type -> gw.RidReq
	17, // 9: gw.GatewayI.rids:input_type -> gw.RidsReq
	12, // 10: gw.GatewayI.cids:input_type -> gw.GidReq
	1,  // 11: gw.GatewayI.cidGid:input_type -> gw.CidGidReq
	1,  // 12: gw.GatewayI.gidCid:input_type -> gw.CidGidReq
	18, // 13: gw.GatewayI.conn:input_type -> gw.GConnReq
	19, // 14: gw.GatewayI.disc:input_type -> gw.GDiscReq
	2,  // 15: gw.GatewayI.rep:input_type -> gw.RepReq
	4,  // 16: gw.GatewayI.last:input_type -> gw.ILastReq
	20, // 17: gw.GatewayI.push:input_type -> gw.PushReq
	5,  // 18: gw.GatewayI.gQueue:input_type -> gw.IGQueueReq
	6,  // 19: gw.GatewayI.gClear:input_type -> gw.IGClearReq
	21, // 20: gw.GatewayI.gLasts:input_type -> gw.GLastsReq
	12, // 21: gw.GatewayI.gLast:input_type -> gw.GidReq
	22, // 22: gw.GatewayI.gPush:input_type -> gw.GPushReq
	7,  // 23: gw.GatewayI.gPushA:input_type -> gw.IGPushAReq
	23, // 24: gw.GatewayI.send:input_type -> gw.SendReq
	24, // 25: gw.GatewayI.tPush:input_type -> gw.TPushReq
	12, // 26: gw.GatewayI.tDirty:input_type -> gw.GidReq
	12, // 27: gw.GatewayI.tStarts:input_type -> gw.GidReq
	25, // 28: gw.GatewayI.setProds:input_type -> gw.ProdsRep
	8,  // 29: gw.GatewayI.read:input_type -> gw.ReadReq
	9,  // 30: gw.GatewayI.unread:input_type -> gw.UnreadReq
	10, // 31: gw.GatewayI.unreads:input_type -> gw.UnreadReqs
	26, // 32: gw.GatewayI.unreadTids:input_type -> gw.UnreadTids
	27, // 33: gw.GatewayI.uid:output_type -> gw.UIdRep
	28, // 34: gw.GatewayI.online:output_type -> gw.BoolRep
	29, // 35: gw.GatewayI.onlines:output_type -> gw.BoolsRep
	30, // 36: gw.GatewayI.close:output_type -> gw.Id32Rep
	30, // 37: gw.GatewayI.kick:output_type -> gw.Id32Rep
	30, // 38: gw.GatewayI.alive:output_type -> gw.Id32Rep
	30, // 39: gw.GatewayI.rid:output_type -> gw.Id32Rep
	30, // 40: gw.GatewayI.rids:output_type -> gw.Id32Rep
	31, // 41: gw.GatewayI.cids:output_type -> gw.CidsRep
	30, // 42: gw.GatewayI.cidGid:output_type -> gw.Id32Rep
	30, // 43: gw.GatewayI.gidCid:output_type -> gw.Id32Rep
	30, // 44: gw.GatewayI.conn:output_type -> gw.Id32Rep
	30, // 45: gw.GatewayI.disc:output_type -> gw.Id32Rep
	30, // 46: gw.GatewayI.rep:output_type -> gw.Id32Rep
	30, // 47: gw.GatewayI.last:output_type -> gw.Id32Rep
	30, // 48: gw.GatewayI.push:output_type -> gw.Id32Rep
	30, // 49: gw.GatewayI.gQueue:output_type -> gw.Id32Rep
	30, // 50: gw.GatewayI.gClear:output_type -> gw.Id32Rep
	30, // 51: gw.GatewayI.gLasts:output_type -> gw.Id32Rep
	30, // 52: gw.GatewayI.gLast:output_type -> gw.Id32Rep
	32, // 53: gw.GatewayI.gPush:output_type -> gw.Id64Rep
	30, // 54: gw.GatewayI.gPushA:output_type -> gw.Id32Rep
	30, // 55: gw.GatewayI.send:output_type -> gw.Id32Rep
	30, // 56: gw.GatewayI.tPush:output_type -> gw.Id32Rep
	30, // 57: gw.GatewayI.tDirty:output_type -> gw.Id32Rep
	30, // 58: gw.GatewayI.tStarts:output_type -> gw.Id32Rep
	28, // 59: gw.GatewayI.setProds:output_type -> gw.BoolRep
	30, // 60: gw.GatewayI.read:output_type -> gw.Id32Rep
	30, // 61: gw.GatewayI.unread:output_type -> gw.Id32Rep
	30, // 62: gw.GatewayI.unreads:output_type -> gw.Id32Rep
	30, // 63: gw.GatewayI.unreadTids:output_type -> gw.Id32Rep
	33, // [33:64] is the sub-list for method output_type
	2,  // [2:33] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_dsl_gwI_proto_init() }
func file_dsl_gwI_proto_init() {
	if File_dsl_gwI_proto != nil {
		return
	}
	file_dsl_gw_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_dsl_gwI_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CidGidReq); i {
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
		file_dsl_gwI_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepReq); i {
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
		file_dsl_gwI_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Msg); i {
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
		file_dsl_gwI_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ILastReq); i {
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
		file_dsl_gwI_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IGQueueReq); i {
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
		file_dsl_gwI_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IGClearReq); i {
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
		file_dsl_gwI_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IGPushAReq); i {
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
		file_dsl_gwI_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadReq); i {
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
		file_dsl_gwI_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnreadReq); i {
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
		file_dsl_gwI_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnreadReqs); i {
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
			RawDescriptor: file_dsl_gwI_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_dsl_gwI_proto_goTypes,
		DependencyIndexes: file_dsl_gwI_proto_depIdxs,
		EnumInfos:         file_dsl_gwI_proto_enumTypes,
		MessageInfos:      file_dsl_gwI_proto_msgTypes,
	}.Build()
	File_dsl_gwI_proto = out.File
	file_dsl_gwI_proto_rawDesc = nil
	file_dsl_gwI_proto_goTypes = nil
	file_dsl_gwI_proto_depIdxs = nil
}
