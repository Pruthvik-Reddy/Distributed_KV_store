package rpc

import (
	"context"

	"distributed-kv-store/internal/store"
	pb "distributed-kv-store/api/kvstore/v1" // Import generated protobuf code
)

// Server is the implementation of our gRPC server.
// It holds a reference to the key-value store.
type Server struct {
	pb.UnimplementedKVStoreServer // Recommended for forward-compatibility
	store                         store.Store
}

// NewKVServiceServer creates a new instance of our gRPC server.
func NewKVServiceServer(s store.Store) *Server {
	return &Server{store: s}
}

// Put implements the Put RPC method.
// It takes a PutRequest from the client and calls the underlying store's Put method.
func (s *Server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	err := s.store.Put(req.GetKey(), req.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.PutResponse{}, nil
}

// Get implements the Get RPC method.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := s.store.Get(req.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Value: val}, nil
}

// Delete implements the Delete RPC method.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.store.Delete(req.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteResponse{}, nil
}
