// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hello/v1/echo.proto

package hello_v1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// EchoClient is the client API for Echo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EchoClient interface {
	Echo(ctx context.Context, in *HiRequest, opts ...grpc.CallOption) (*HiRequest, error)
}

type echoClient struct {
	cc *grpc.ClientConn
}

func NewEchoClient(cc *grpc.ClientConn) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Echo(ctx context.Context, in *HiRequest, opts ...grpc.CallOption) (*HiRequest, error) {
	out := new(HiRequest)
	err := c.cc.Invoke(ctx, "/hello.v1.Echo/Echo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EchoServer is the server API for Echo service.
type EchoServer interface {
	Echo(context.Context, *HiRequest) (*HiRequest, error)
}

func RegisterEchoServer(s *grpc.Server, srv EchoServer) {
	s.RegisterService(&_Echo_serviceDesc, srv)
}

func _Echo_Echo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HiRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hello.v1.Echo/Echo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Echo(ctx, req.(*HiRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Echo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "hello.v1.Echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler:    _Echo_Echo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hello/v1/echo.proto",
}

func init() { proto.RegisterFile("hello/v1/echo.proto", fileDescriptor_echo_78742650fb6da331) }

var fileDescriptor_echo_78742650fb6da331 = []byte{
	// 93 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xce, 0x48, 0xcd, 0xc9,
	0xc9, 0xd7, 0x2f, 0x33, 0xd4, 0x4f, 0x4d, 0xce, 0xc8, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0xe2, 0x00, 0x0b, 0xea, 0x95, 0x19, 0x4a, 0x89, 0xc0, 0xa5, 0x21, 0x42, 0x60, 0x79, 0x23, 0x0b,
	0x2e, 0x16, 0xd7, 0xe4, 0x8c, 0x7c, 0x21, 0x03, 0x28, 0x2d, 0xac, 0x07, 0xd3, 0xa0, 0xe7, 0x91,
	0x19, 0x94, 0x5a, 0x58, 0x9a, 0x5a, 0x5c, 0x22, 0x85, 0x4d, 0x30, 0x89, 0x0d, 0x6c, 0x80, 0x31,
	0x20, 0x00, 0x00, 0xff, 0xff, 0x36, 0x25, 0xeb, 0x45, 0x77, 0x00, 0x00, 0x00,
}
