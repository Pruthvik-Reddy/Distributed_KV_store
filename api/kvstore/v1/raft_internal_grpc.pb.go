// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v6.31.1
// source: api/kvstore/v1/raft_internal.proto

package kvstore

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

// RaftInternalClient is the client API for RaftInternal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftInternalClient interface {
	// RequestVote is called by candidates to gather votes.
	RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error)
	// AppendEntries is called by leader to replicate log entries; also used as heartbeat.
	AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error)
}

type raftInternalClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftInternalClient(cc grpc.ClientConnInterface) RaftInternalClient {
	return &raftInternalClient{cc}
}

func (c *raftInternalClient) RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error) {
	out := new(RequestVoteReply)
	err := c.cc.Invoke(ctx, "/kvstore.v1.RaftInternal/RequestVote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftInternalClient) AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error) {
	out := new(AppendEntriesReply)
	err := c.cc.Invoke(ctx, "/kvstore.v1.RaftInternal/AppendEntries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftInternalServer is the server API for RaftInternal service.
// All implementations must embed UnimplementedRaftInternalServer
// for forward compatibility
type RaftInternalServer interface {
	// RequestVote is called by candidates to gather votes.
	RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error)
	// AppendEntries is called by leader to replicate log entries; also used as heartbeat.
	AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error)
	mustEmbedUnimplementedRaftInternalServer()
}

// UnimplementedRaftInternalServer must be embedded to have forward compatible implementations.
type UnimplementedRaftInternalServer struct {
}

func (UnimplementedRaftInternalServer) RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedRaftInternalServer) AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftInternalServer) mustEmbedUnimplementedRaftInternalServer() {}

// UnsafeRaftInternalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftInternalServer will
// result in compilation errors.
type UnsafeRaftInternalServer interface {
	mustEmbedUnimplementedRaftInternalServer()
}

func RegisterRaftInternalServer(s grpc.ServiceRegistrar, srv RaftInternalServer) {
	s.RegisterService(&RaftInternal_ServiceDesc, srv)
}

func _RaftInternal_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestVoteArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftInternalServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kvstore.v1.RaftInternal/RequestVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftInternalServer).RequestVote(ctx, req.(*RequestVoteArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftInternal_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftInternalServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kvstore.v1.RaftInternal/AppendEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftInternalServer).AppendEntries(ctx, req.(*AppendEntriesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftInternal_ServiceDesc is the grpc.ServiceDesc for RaftInternal service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftInternal_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kvstore.v1.RaftInternal",
	HandlerType: (*RaftInternalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestVote",
			Handler:    _RaftInternal_RequestVote_Handler,
		},
		{
			MethodName: "AppendEntries",
			Handler:    _RaftInternal_AppendEntries_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/kvstore/v1/raft_internal.proto",
}
