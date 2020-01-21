// Code generated by protoc-gen-go. DO NOT EDIT.
// source: payments.proto

package payments

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

type ChargeReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChargeReq) Reset()         { *m = ChargeReq{} }
func (m *ChargeReq) String() string { return proto.CompactTextString(m) }
func (*ChargeReq) ProtoMessage()    {}
func (*ChargeReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9566e6e864d2854, []int{0}
}

func (m *ChargeReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChargeReq.Unmarshal(m, b)
}
func (m *ChargeReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChargeReq.Marshal(b, m, deterministic)
}
func (m *ChargeReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChargeReq.Merge(m, src)
}
func (m *ChargeReq) XXX_Size() int {
	return xxx_messageInfo_ChargeReq.Size(m)
}
func (m *ChargeReq) XXX_DiscardUnknown() {
	xxx_messageInfo_ChargeReq.DiscardUnknown(m)
}

var xxx_messageInfo_ChargeReq proto.InternalMessageInfo

type ChargeReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChargeReply) Reset()         { *m = ChargeReply{} }
func (m *ChargeReply) String() string { return proto.CompactTextString(m) }
func (*ChargeReply) ProtoMessage()    {}
func (*ChargeReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9566e6e864d2854, []int{1}
}

func (m *ChargeReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChargeReply.Unmarshal(m, b)
}
func (m *ChargeReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChargeReply.Marshal(b, m, deterministic)
}
func (m *ChargeReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChargeReply.Merge(m, src)
}
func (m *ChargeReply) XXX_Size() int {
	return xxx_messageInfo_ChargeReply.Size(m)
}
func (m *ChargeReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ChargeReply.DiscardUnknown(m)
}

var xxx_messageInfo_ChargeReply proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ChargeReq)(nil), "payments.ChargeReq")
	proto.RegisterType((*ChargeReply)(nil), "payments.ChargeReply")
}

func init() { proto.RegisterFile("payments.proto", fileDescriptor_a9566e6e864d2854) }

var fileDescriptor_a9566e6e864d2854 = []byte{
	// 98 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2b, 0x48, 0xac, 0xcc,
	0x4d, 0xcd, 0x2b, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0x95, 0xb8,
	0xb9, 0x38, 0x9d, 0x33, 0x12, 0x8b, 0xd2, 0x53, 0x83, 0x52, 0x0b, 0x95, 0x78, 0xb9, 0xb8, 0x61,
	0x9c, 0x82, 0x9c, 0x4a, 0x23, 0x67, 0x2e, 0x8e, 0x00, 0xa8, 0x3a, 0x21, 0x73, 0xb8, 0xba, 0x00,
	0x67, 0x21, 0x61, 0x3d, 0xb8, 0x79, 0x70, 0xcd, 0x52, 0xa2, 0x98, 0x82, 0x05, 0x39, 0x95, 0x49,
	0x6c, 0x60, 0x1b, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x40, 0x40, 0x4e, 0xfe, 0x83, 0x00,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PaymentsClient is the client API for Payments service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PaymentsClient interface {
	ChargeRPC(ctx context.Context, in *ChargeReq, opts ...grpc.CallOption) (*ChargeReply, error)
}

type paymentsClient struct {
	cc *grpc.ClientConn
}

func NewPaymentsClient(cc *grpc.ClientConn) PaymentsClient {
	return &paymentsClient{cc}
}

func (c *paymentsClient) ChargeRPC(ctx context.Context, in *ChargeReq, opts ...grpc.CallOption) (*ChargeReply, error) {
	out := new(ChargeReply)
	err := c.cc.Invoke(ctx, "/payments.Payments/ChargeRPC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PaymentsServer is the server API for Payments service.
type PaymentsServer interface {
	ChargeRPC(context.Context, *ChargeReq) (*ChargeReply, error)
}

// UnimplementedPaymentsServer can be embedded to have forward compatible implementations.
type UnimplementedPaymentsServer struct {
}

func (*UnimplementedPaymentsServer) ChargeRPC(ctx context.Context, req *ChargeReq) (*ChargeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChargeRPC not implemented")
}

func RegisterPaymentsServer(s *grpc.Server, srv PaymentsServer) {
	s.RegisterService(&_Payments_serviceDesc, srv)
}

func _Payments_ChargeRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChargeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentsServer).ChargeRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payments.Payments/ChargeRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentsServer).ChargeRPC(ctx, req.(*ChargeReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Payments_serviceDesc = grpc.ServiceDesc{
	ServiceName: "payments.Payments",
	HandlerType: (*PaymentsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChargeRPC",
			Handler:    _Payments_ChargeRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payments.proto",
}