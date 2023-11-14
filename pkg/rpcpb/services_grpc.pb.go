// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: pkg/rpcpb/services.proto

package rpcpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TranslatorClient is the client API for Translator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TranslatorClient interface {
	TranslateTo(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Data, error)
	TranslateFrom(ctx context.Context, in *Data, opts ...grpc.CallOption) (*Reply, error)
	GetCmdDescriptions(ctx context.Context, in *DescriptionsRequest, opts ...grpc.CallOption) (*DescriptionsReply, error)
}

type translatorClient struct {
	cc grpc.ClientConnInterface
}

func NewTranslatorClient(cc grpc.ClientConnInterface) TranslatorClient {
	return &translatorClient{cc}
}

func (c *translatorClient) TranslateTo(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Data, error) {
	out := new(Data)
	err := c.cc.Invoke(ctx, "/rpcpb.Translator/TranslateTo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translatorClient) TranslateFrom(ctx context.Context, in *Data, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/rpcpb.Translator/TranslateFrom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translatorClient) GetCmdDescriptions(ctx context.Context, in *DescriptionsRequest, opts ...grpc.CallOption) (*DescriptionsReply, error) {
	out := new(DescriptionsReply)
	err := c.cc.Invoke(ctx, "/rpcpb.Translator/GetCmdDescriptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TranslatorServer is the server API for Translator service.
// All implementations must embed UnimplementedTranslatorServer
// for forward compatibility
type TranslatorServer interface {
	TranslateTo(context.Context, *Request) (*Data, error)
	TranslateFrom(context.Context, *Data) (*Reply, error)
	GetCmdDescriptions(context.Context, *DescriptionsRequest) (*DescriptionsReply, error)
	mustEmbedUnimplementedTranslatorServer()
}

// UnimplementedTranslatorServer must be embedded to have forward compatible implementations.
type UnimplementedTranslatorServer struct {
}

func (UnimplementedTranslatorServer) TranslateTo(context.Context, *Request) (*Data, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TranslateTo not implemented")
}
func (UnimplementedTranslatorServer) TranslateFrom(context.Context, *Data) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TranslateFrom not implemented")
}
func (UnimplementedTranslatorServer) GetCmdDescriptions(context.Context, *DescriptionsRequest) (*DescriptionsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCmdDescriptions not implemented")
}
func (UnimplementedTranslatorServer) mustEmbedUnimplementedTranslatorServer() {}

// UnsafeTranslatorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TranslatorServer will
// result in compilation errors.
type UnsafeTranslatorServer interface {
	mustEmbedUnimplementedTranslatorServer()
}

func RegisterTranslatorServer(s grpc.ServiceRegistrar, srv TranslatorServer) {
	s.RegisterService(&Translator_ServiceDesc, srv)
}

func _Translator_TranslateTo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslatorServer).TranslateTo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcpb.Translator/TranslateTo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslatorServer).TranslateTo(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Translator_TranslateFrom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Data)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslatorServer).TranslateFrom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcpb.Translator/TranslateFrom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslatorServer).TranslateFrom(ctx, req.(*Data))
	}
	return interceptor(ctx, in, info, handler)
}

func _Translator_GetCmdDescriptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DescriptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslatorServer).GetCmdDescriptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcpb.Translator/GetCmdDescriptions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslatorServer).GetCmdDescriptions(ctx, req.(*DescriptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Translator_ServiceDesc is the grpc.ServiceDesc for Translator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Translator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpcpb.Translator",
	HandlerType: (*TranslatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TranslateTo",
			Handler:    _Translator_TranslateTo_Handler,
		},
		{
			MethodName: "TranslateFrom",
			Handler:    _Translator_TranslateFrom_Handler,
		},
		{
			MethodName: "GetCmdDescriptions",
			Handler:    _Translator_GetCmdDescriptions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/rpcpb/services.proto",
}

// BuilderClient is the client API for Builder service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BuilderClient interface {
	GetParams(ctx context.Context, in *ParamsRequest, opts ...grpc.CallOption) (*ParamsReply, error)
	BuildAgent(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (*BuildReply, error)
}

type builderClient struct {
	cc grpc.ClientConnInterface
}

func NewBuilderClient(cc grpc.ClientConnInterface) BuilderClient {
	return &builderClient{cc}
}

func (c *builderClient) GetParams(ctx context.Context, in *ParamsRequest, opts ...grpc.CallOption) (*ParamsReply, error) {
	out := new(ParamsReply)
	err := c.cc.Invoke(ctx, "/rpcpb.Builder/GetParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *builderClient) BuildAgent(ctx context.Context, in *BuildRequest, opts ...grpc.CallOption) (*BuildReply, error) {
	out := new(BuildReply)
	err := c.cc.Invoke(ctx, "/rpcpb.Builder/BuildAgent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuilderServer is the server API for Builder service.
// All implementations must embed UnimplementedBuilderServer
// for forward compatibility
type BuilderServer interface {
	GetParams(context.Context, *ParamsRequest) (*ParamsReply, error)
	BuildAgent(context.Context, *BuildRequest) (*BuildReply, error)
	mustEmbedUnimplementedBuilderServer()
}

// UnimplementedBuilderServer must be embedded to have forward compatible implementations.
type UnimplementedBuilderServer struct {
}

func (UnimplementedBuilderServer) GetParams(context.Context, *ParamsRequest) (*ParamsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetParams not implemented")
}
func (UnimplementedBuilderServer) BuildAgent(context.Context, *BuildRequest) (*BuildReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildAgent not implemented")
}
func (UnimplementedBuilderServer) mustEmbedUnimplementedBuilderServer() {}

// UnsafeBuilderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BuilderServer will
// result in compilation errors.
type UnsafeBuilderServer interface {
	mustEmbedUnimplementedBuilderServer()
}

func RegisterBuilderServer(s grpc.ServiceRegistrar, srv BuilderServer) {
	s.RegisterService(&Builder_ServiceDesc, srv)
}

func _Builder_GetParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderServer).GetParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcpb.Builder/GetParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderServer).GetParams(ctx, req.(*ParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Builder_BuildAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderServer).BuildAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcpb.Builder/BuildAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderServer).BuildAgent(ctx, req.(*BuildRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Builder_ServiceDesc is the grpc.ServiceDesc for Builder service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Builder_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpcpb.Builder",
	HandlerType: (*BuilderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetParams",
			Handler:    _Builder_GetParams_Handler,
		},
		{
			MethodName: "BuildAgent",
			Handler:    _Builder_BuildAgent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/rpcpb/services.proto",
}
