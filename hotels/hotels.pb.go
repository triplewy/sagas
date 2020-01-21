// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hotels.proto

package hotels

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type BookReq struct {
	UserID               string   `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	RoomID               string   `protobuf:"bytes,2,opt,name=roomID,proto3" json:"roomID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BookReq) Reset()         { *m = BookReq{} }
func (m *BookReq) String() string { return proto.CompactTextString(m) }
func (*BookReq) ProtoMessage()    {}
func (*BookReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_f848554aaca8d36d, []int{0}
}

func (m *BookReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BookReq.Unmarshal(m, b)
}
func (m *BookReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BookReq.Marshal(b, m, deterministic)
}
func (m *BookReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BookReq.Merge(m, src)
}
func (m *BookReq) XXX_Size() int {
	return xxx_messageInfo_BookReq.Size(m)
}
func (m *BookReq) XXX_DiscardUnknown() {
	xxx_messageInfo_BookReq.DiscardUnknown(m)
}

var xxx_messageInfo_BookReq proto.InternalMessageInfo

func (m *BookReq) GetUserID() string {
	if m != nil {
		return m.UserID
	}
	return ""
}

func (m *BookReq) GetRoomID() string {
	if m != nil {
		return m.RoomID
	}
	return ""
}

type BookReply struct {
	ReservationID        string   `protobuf:"bytes,1,opt,name=reservationID,proto3" json:"reservationID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BookReply) Reset()         { *m = BookReply{} }
func (m *BookReply) String() string { return proto.CompactTextString(m) }
func (*BookReply) ProtoMessage()    {}
func (*BookReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_f848554aaca8d36d, []int{1}
}

func (m *BookReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BookReply.Unmarshal(m, b)
}
func (m *BookReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BookReply.Marshal(b, m, deterministic)
}
func (m *BookReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BookReply.Merge(m, src)
}
func (m *BookReply) XXX_Size() int {
	return xxx_messageInfo_BookReply.Size(m)
}
func (m *BookReply) XXX_DiscardUnknown() {
	xxx_messageInfo_BookReply.DiscardUnknown(m)
}

var xxx_messageInfo_BookReply proto.InternalMessageInfo

func (m *BookReply) GetReservationID() string {
	if m != nil {
		return m.ReservationID
	}
	return ""
}

type CancelReq struct {
	UserID               string   `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	ReservationID        string   `protobuf:"bytes,2,opt,name=reservationID,proto3" json:"reservationID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CancelReq) Reset()         { *m = CancelReq{} }
func (m *CancelReq) String() string { return proto.CompactTextString(m) }
func (*CancelReq) ProtoMessage()    {}
func (*CancelReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_f848554aaca8d36d, []int{2}
}

func (m *CancelReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelReq.Unmarshal(m, b)
}
func (m *CancelReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelReq.Marshal(b, m, deterministic)
}
func (m *CancelReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelReq.Merge(m, src)
}
func (m *CancelReq) XXX_Size() int {
	return xxx_messageInfo_CancelReq.Size(m)
}
func (m *CancelReq) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelReq.DiscardUnknown(m)
}

var xxx_messageInfo_CancelReq proto.InternalMessageInfo

func (m *CancelReq) GetUserID() string {
	if m != nil {
		return m.UserID
	}
	return ""
}

func (m *CancelReq) GetReservationID() string {
	if m != nil {
		return m.ReservationID
	}
	return ""
}

type CancelReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CancelReply) Reset()         { *m = CancelReply{} }
func (m *CancelReply) String() string { return proto.CompactTextString(m) }
func (*CancelReply) ProtoMessage()    {}
func (*CancelReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_f848554aaca8d36d, []int{3}
}

func (m *CancelReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelReply.Unmarshal(m, b)
}
func (m *CancelReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelReply.Marshal(b, m, deterministic)
}
func (m *CancelReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelReply.Merge(m, src)
}
func (m *CancelReply) XXX_Size() int {
	return xxx_messageInfo_CancelReply.Size(m)
}
func (m *CancelReply) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelReply.DiscardUnknown(m)
}

var xxx_messageInfo_CancelReply proto.InternalMessageInfo

func init() {
	proto.RegisterType((*BookReq)(nil), "hotels.BookReq")
	proto.RegisterType((*BookReply)(nil), "hotels.BookReply")
	proto.RegisterType((*CancelReq)(nil), "hotels.CancelReq")
	proto.RegisterType((*CancelReply)(nil), "hotels.CancelReply")
}

func init() { proto.RegisterFile("hotels.proto", fileDescriptor_f848554aaca8d36d) }

var fileDescriptor_f848554aaca8d36d = []byte{
	// 189 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xc9, 0xc8, 0x2f, 0x49,
	0xcd, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x83, 0xf0, 0x94, 0x2c, 0xb9, 0xd8,
	0x9d, 0xf2, 0xf3, 0xb3, 0x83, 0x52, 0x0b, 0x85, 0xc4, 0xb8, 0xd8, 0x4a, 0x8b, 0x53, 0x8b, 0x3c,
	0x5d, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xa0, 0x3c, 0x90, 0x78, 0x51, 0x7e, 0x7e, 0xae,
	0xa7, 0x8b, 0x04, 0x13, 0x44, 0x1c, 0xc2, 0x53, 0x32, 0xe4, 0xe2, 0x84, 0x68, 0x2d, 0xc8, 0xa9,
	0x14, 0x52, 0xe1, 0xe2, 0x2d, 0x4a, 0x2d, 0x4e, 0x2d, 0x2a, 0x4b, 0x2c, 0xc9, 0xcc, 0xcf, 0x83,
	0x9b, 0x81, 0x2a, 0xa8, 0xe4, 0xc9, 0xc5, 0xe9, 0x9c, 0x98, 0x97, 0x9c, 0x9a, 0x83, 0xcf, 0x3e,
	0x0c, 0xa3, 0x98, 0xb0, 0x19, 0xc5, 0xcb, 0xc5, 0x0d, 0x33, 0xaa, 0x20, 0xa7, 0xd2, 0x28, 0x87,
	0x8b, 0xcd, 0x03, 0xec, 0x23, 0x21, 0x5d, 0xa8, 0x8f, 0x02, 0x9c, 0x85, 0xf8, 0xf5, 0xa0, 0x7e,
	0x86, 0x7a, 0x51, 0x4a, 0x10, 0x55, 0x00, 0xe4, 0x70, 0x63, 0xb8, 0x93, 0x02, 0x9c, 0x85, 0xe0,
	0xf2, 0x70, 0x57, 0x4a, 0x09, 0xa3, 0x0b, 0x15, 0xe4, 0x54, 0x26, 0xb1, 0x81, 0x03, 0xd1, 0x18,
	0x10, 0x00, 0x00, 0xff, 0xff, 0x04, 0x73, 0x38, 0x20, 0x54, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// HotelsClient is the client API for Hotels service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HotelsClient interface {
	BookRPC(ctx context.Context, in *BookReq, opts ...grpc.CallOption) (*BookReply, error)
	CancelRPC(ctx context.Context, in *CancelReq, opts ...grpc.CallOption) (*CancelReply, error)
}

type hotelsClient struct {
	cc *grpc.ClientConn
}

func NewHotelsClient(cc *grpc.ClientConn) HotelsClient {
	return &hotelsClient{cc}
}

func (c *hotelsClient) BookRPC(ctx context.Context, in *BookReq, opts ...grpc.CallOption) (*BookReply, error) {
	out := new(BookReply)
	err := c.cc.Invoke(ctx, "/hotels.Hotels/BookRPC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hotelsClient) CancelRPC(ctx context.Context, in *CancelReq, opts ...grpc.CallOption) (*CancelReply, error) {
	out := new(CancelReply)
	err := c.cc.Invoke(ctx, "/hotels.Hotels/CancelRPC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HotelsServer is the server API for Hotels service.
type HotelsServer interface {
	BookRPC(context.Context, *BookReq) (*BookReply, error)
	CancelRPC(context.Context, *CancelReq) (*CancelReply, error)
}

// UnimplementedHotelsServer can be embedded to have forward compatible implementations.
type UnimplementedHotelsServer struct {
}

func (*UnimplementedHotelsServer) BookRPC(ctx context.Context, req *BookReq) (*BookReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BookRPC not implemented")
}
func (*UnimplementedHotelsServer) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelRPC not implemented")
}

func RegisterHotelsServer(s *grpc.Server, srv HotelsServer) {
	s.RegisterService(&_Hotels_serviceDesc, srv)
}

func _Hotels_BookRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BookReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HotelsServer).BookRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hotels.Hotels/BookRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HotelsServer).BookRPC(ctx, req.(*BookReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hotels_CancelRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HotelsServer).CancelRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hotels.Hotels/CancelRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HotelsServer).CancelRPC(ctx, req.(*CancelReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Hotels_serviceDesc = grpc.ServiceDesc{
	ServiceName: "hotels.Hotels",
	HandlerType: (*HotelsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BookRPC",
			Handler:    _Hotels_BookRPC_Handler,
		},
		{
			MethodName: "CancelRPC",
			Handler:    _Hotels_CancelRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hotels.proto",
}