package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	// Our internal packages
	"distributed-kv-store/internal/rpc"
	"distributed-kv-store/internal/store"

	// Generated protobuf package
	pb "distributed-kv-store/api/kvstore/v1"

	"google.golang.org/grpc"
)

func main() {
	// Define the network address for the gRPC server.
	listenAddr := ":50051"
	log.Printf("Starting gRPC server on %s", listenAddr)

	// Create a TCP listener on the specified address.
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Initialize our key-value store.
	dbPath := "./data"
	kvStore, err := store.NewLevelDBStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer kvStore.Close()

	// Create a new gRPC server instance.
	gRPCServer := grpc.NewServer()

	// Create an instance of our KV service implementation.
	kvServiceServer := rpc.NewKVServiceServer(kvStore)

	// Register our service implementation with the gRPC server.
	pb.RegisterKVStoreServer(gRPCServer, kvServiceServer)

	// Start the gRPC server in a separate goroutine so it doesn't block.
	go func() {
		if err := gRPCServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
	log.Printf("Server is ready to accept connections.")

	// --- Graceful Shutdown ---
	// Wait for a shutdown signal (e.g., Ctrl+C).
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh // Block until a signal is received.

	log.Println("Shutdown signal received, gracefully stopping gRPC server...")
	gRPCServer.GracefulStop()
	log.Println("Server stopped.")
}