package rpc

import (
	"context"
	"fmt"
	

	"distributed-kv-store/internal/raft" // Import raft package
	"distributed-kv-store/internal/store"
	pb "distributed-kv-store/api/kvstore/v1"
)
import "distributed-kv-store/internal/monitoring"
type Server struct {
	pb.UnimplementedKVStoreServer
	store    store.Store
	raftNode *raft.Node // Add a reference to the Raft node
}

// NewKVServiceServer now also takes a raft.Node as a dependency.
func NewKVServiceServer(s store.Store, r *raft.Node) *Server {
	return &Server{
		store:    s,
		raftNode: r,
	}
}

// Put now proposes the command to Raft instead of writing directly to the store.
func (s *Server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	// Simple command serialization: "PUT:key:value"
	monitoring.RequestsTotal.WithLabelValues("put").Inc()
	cmd := []byte(fmt.Sprintf("PUT:%s:%s", req.GetKey(), req.GetValue()))
	
	// Propose the command to the Raft cluster.
	_, _, isLeader := s.raftNode.Propose(cmd)
	
	if !isLeader {
		return nil, fmt.Errorf("this node is not the leader")
	}

	// For a real implementation, you would wait for the command to be applied.
	// For this learning step, we'll assume it will be applied and return success.
	return &pb.PutResponse{}, nil
}

// Get still reads from the local store, but we should check for leadership.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// For strong consistency, reads should also go through the leader.
	// We can check leadership status, but Raft doesn't expose it directly in this simple model.
	// A simple workaround is to try a lightweight proposal or check a status method.
	// For now, we'll just read from the local store. In a real system, this could be stale.
	monitoring.RequestsTotal.WithLabelValues("get").Inc()
	val, err := s.store.Get(req.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Value: val}, nil
}

// Delete now proposes the command to Raft.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	monitoring.RequestsTotal.WithLabelValues("delete").Inc()
	cmd := []byte(fmt.Sprintf("DEL:%s", req.GetKey()))

	_, _, isLeader := s.raftNode.Propose(cmd)
	if !isLeader {
		return nil, fmt.Errorf("this node is not the leader")
	}

	return &pb.DeleteResponse{}, nil
}