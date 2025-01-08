// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.1
// source: connector/service.proto

package connector

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
	ConnectorService_SendMessage_FullMethodName = "/ConnectorService/SendMessage"
	ConnectorService_GetVersion_FullMethodName  = "/ConnectorService/GetVersion"
)

// ConnectorServiceClient is the client API for ConnectorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ConnectorServiceClient interface {
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error)
}

type connectorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewConnectorServiceClient(cc grpc.ClientConnInterface) ConnectorServiceClient {
	return &connectorServiceClient{cc}
}

func (c *connectorServiceClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, ConnectorService_SendMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *connectorServiceClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, ConnectorService_GetVersion_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConnectorServiceServer is the server API for ConnectorService service.
// All implementations must embed UnimplementedConnectorServiceServer
// for forward compatibility.
type ConnectorServiceServer interface {
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error)
	mustEmbedUnimplementedConnectorServiceServer()
}

// UnimplementedConnectorServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedConnectorServiceServer struct{}

func (UnimplementedConnectorServiceServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedConnectorServiceServer) GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedConnectorServiceServer) mustEmbedUnimplementedConnectorServiceServer() {}
func (UnimplementedConnectorServiceServer) testEmbeddedByValue()                          {}

// UnsafeConnectorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ConnectorServiceServer will
// result in compilation errors.
type UnsafeConnectorServiceServer interface {
	mustEmbedUnimplementedConnectorServiceServer()
}

func RegisterConnectorServiceServer(s grpc.ServiceRegistrar, srv ConnectorServiceServer) {
	// If the following call pancis, it indicates UnimplementedConnectorServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ConnectorService_ServiceDesc, srv)
}

func _ConnectorService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConnectorService_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConnectorServiceServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConnectorService_GetVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConnectorServiceServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ConnectorService_ServiceDesc is the grpc.ServiceDesc for ConnectorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ConnectorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ConnectorService",
	HandlerType: (*ConnectorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _ConnectorService_SendMessage_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _ConnectorService_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "connector/service.proto",
}
