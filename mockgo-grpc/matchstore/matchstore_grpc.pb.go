// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.15.8
// source: matchstore/matchstore.proto

package matchstore

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

// MatchstoreClient is the client API for Matchstore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MatchstoreClient interface {
	FetchMatches(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*MatchesResponse, error)
	FetchMismatches(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*MismatchesResponse, error)
	FetchMatchesCount(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*MatchesCountResponse, error)
	FetchMismatchesCount(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*MismatchesCountResponse, error)
	RemoveMatches(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*RemoveResponse, error)
	RemoveMismatches(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*RemoveResponse, error)
}

type matchstoreClient struct {
	cc grpc.ClientConnInterface
}

func NewMatchstoreClient(cc grpc.ClientConnInterface) MatchstoreClient {
	return &matchstoreClient{cc}
}

func (c *matchstoreClient) FetchMatches(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*MatchesResponse, error) {
	out := new(MatchesResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/FetchMatches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matchstoreClient) FetchMismatches(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*MismatchesResponse, error) {
	out := new(MismatchesResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/FetchMismatches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matchstoreClient) FetchMatchesCount(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*MatchesCountResponse, error) {
	out := new(MatchesCountResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/FetchMatchesCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matchstoreClient) FetchMismatchesCount(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*MismatchesCountResponse, error) {
	out := new(MismatchesCountResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/FetchMismatchesCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matchstoreClient) RemoveMatches(ctx context.Context, in *EndPointRequest, opts ...grpc.CallOption) (*RemoveResponse, error) {
	out := new(RemoveResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/RemoveMatches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matchstoreClient) RemoveMismatches(ctx context.Context, in *MismatchRequest, opts ...grpc.CallOption) (*RemoveResponse, error) {
	out := new(RemoveResponse)
	err := c.cc.Invoke(ctx, "/matchstore.Matchstore/RemoveMismatches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MatchstoreServer is the server API for Matchstore service.
// All implementations must embed UnimplementedMatchstoreServer
// for forward compatibility
type MatchstoreServer interface {
	FetchMatches(context.Context, *EndPointRequest) (*MatchesResponse, error)
	FetchMismatches(context.Context, *MismatchRequest) (*MismatchesResponse, error)
	FetchMatchesCount(context.Context, *EndPointRequest) (*MatchesCountResponse, error)
	FetchMismatchesCount(context.Context, *MismatchRequest) (*MismatchesCountResponse, error)
	RemoveMatches(context.Context, *EndPointRequest) (*RemoveResponse, error)
	RemoveMismatches(context.Context, *MismatchRequest) (*RemoveResponse, error)
	mustEmbedUnimplementedMatchstoreServer()
}

// UnimplementedMatchstoreServer must be embedded to have forward compatible implementations.
type UnimplementedMatchstoreServer struct {
}

func (UnimplementedMatchstoreServer) FetchMatches(context.Context, *EndPointRequest) (*MatchesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMatches not implemented")
}
func (UnimplementedMatchstoreServer) FetchMismatches(context.Context, *MismatchRequest) (*MismatchesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMismatches not implemented")
}
func (UnimplementedMatchstoreServer) FetchMatchesCount(context.Context, *EndPointRequest) (*MatchesCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMatchesCount not implemented")
}
func (UnimplementedMatchstoreServer) FetchMismatchesCount(context.Context, *MismatchRequest) (*MismatchesCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMismatchesCount not implemented")
}
func (UnimplementedMatchstoreServer) RemoveMatches(context.Context, *EndPointRequest) (*RemoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveMatches not implemented")
}
func (UnimplementedMatchstoreServer) RemoveMismatches(context.Context, *MismatchRequest) (*RemoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveMismatches not implemented")
}
func (UnimplementedMatchstoreServer) mustEmbedUnimplementedMatchstoreServer() {}

// UnsafeMatchstoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MatchstoreServer will
// result in compilation errors.
type UnsafeMatchstoreServer interface {
	mustEmbedUnimplementedMatchstoreServer()
}

func RegisterMatchstoreServer(s grpc.ServiceRegistrar, srv MatchstoreServer) {
	s.RegisterService(&Matchstore_ServiceDesc, srv)
}

func _Matchstore_FetchMatches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).FetchMatches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/FetchMatches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).FetchMatches(ctx, req.(*EndPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matchstore_FetchMismatches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MismatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).FetchMismatches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/FetchMismatches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).FetchMismatches(ctx, req.(*MismatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matchstore_FetchMatchesCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).FetchMatchesCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/FetchMatchesCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).FetchMatchesCount(ctx, req.(*EndPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matchstore_FetchMismatchesCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MismatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).FetchMismatchesCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/FetchMismatchesCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).FetchMismatchesCount(ctx, req.(*MismatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matchstore_RemoveMatches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).RemoveMatches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/RemoveMatches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).RemoveMatches(ctx, req.(*EndPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matchstore_RemoveMismatches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MismatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatchstoreServer).RemoveMismatches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matchstore.Matchstore/RemoveMismatches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatchstoreServer).RemoveMismatches(ctx, req.(*MismatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Matchstore_ServiceDesc is the grpc.ServiceDesc for Matchstore service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Matchstore_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "matchstore.Matchstore",
	HandlerType: (*MatchstoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchMatches",
			Handler:    _Matchstore_FetchMatches_Handler,
		},
		{
			MethodName: "FetchMismatches",
			Handler:    _Matchstore_FetchMismatches_Handler,
		},
		{
			MethodName: "FetchMatchesCount",
			Handler:    _Matchstore_FetchMatchesCount_Handler,
		},
		{
			MethodName: "FetchMismatchesCount",
			Handler:    _Matchstore_FetchMismatchesCount_Handler,
		},
		{
			MethodName: "RemoveMatches",
			Handler:    _Matchstore_RemoveMatches_Handler,
		},
		{
			MethodName: "RemoveMismatches",
			Handler:    _Matchstore_RemoveMismatches_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "matchstore/matchstore.proto",
}
