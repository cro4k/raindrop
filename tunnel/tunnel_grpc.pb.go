// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.1
// source: tunnel/tunnel.proto

package tunnel

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	RaindropTunnel_SendMessage_FullMethodName = "/RaindropTunnel/SendMessage"
	RaindropTunnel_GetVersion_FullMethodName  = "/RaindropTunnel/GetVersion"
)

// RaindropTunnelClient is the client API for RaindropTunnel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaindropTunnelClient interface {
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error)
}

type raindropTunnelClient struct {
	cc grpc.ClientConnInterface
}

func NewRaindropTunnelClient(cc grpc.ClientConnInterface) RaindropTunnelClient {
	return &raindropTunnelClient{cc}
}

func (c *raindropTunnelClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, RaindropTunnel_SendMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raindropTunnelClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, RaindropTunnel_GetVersion_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaindropTunnelServer is the server API for RaindropTunnel service.
// All implementations must embed UnimplementedRaindropTunnelServer
// for forward compatibility.
type RaindropTunnelServer interface {
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error)
	mustEmbedUnimplementedRaindropTunnelServer()
}

// UnimplementedRaindropTunnelServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRaindropTunnelServer struct{}

func (UnimplementedRaindropTunnelServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedRaindropTunnelServer) GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedRaindropTunnelServer) mustEmbedUnimplementedRaindropTunnelServer() {}
func (UnimplementedRaindropTunnelServer) testEmbeddedByValue()                        {}

// UnsafeRaindropTunnelServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaindropTunnelServer will
// result in compilation errors.
type UnsafeRaindropTunnelServer interface {
	mustEmbedUnimplementedRaindropTunnelServer()
}

func RegisterRaindropTunnelServer(s grpc.ServiceRegistrar, srv RaindropTunnelServer) {
	// If the following call pancis, it indicates UnimplementedRaindropTunnelServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RaindropTunnel_ServiceDesc, srv)
}

func _RaindropTunnel_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaindropTunnelServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaindropTunnel_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaindropTunnelServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaindropTunnel_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaindropTunnelServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaindropTunnel_GetVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaindropTunnelServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RaindropTunnel_ServiceDesc is the grpc.ServiceDesc for RaindropTunnel service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaindropTunnel_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "RaindropTunnel",
	HandlerType: (*RaindropTunnelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _RaindropTunnel_SendMessage_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _RaindropTunnel_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tunnel/tunnel.proto",
}
