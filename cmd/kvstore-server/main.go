// File: cmd/kvstore-server/main.go
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"distributed-kv-store/internal/raft"
	"distributed-kv-store/internal/rpc"
	"distributed-kv-store/internal/store"

	pb "distributed-kv-store/api/kvstore/v1"

	"google.golang.org/grpc"
)

type stringslice []string
func (s *stringslice) String() string { return strings.Join(*s, ", ") }
func (s *stringslice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var id, kvAddr, raftAddr string
	var peers stringslice
	flag.StringVar(&id, "id", "", "Node ID (required)")
	flag.StringVar(&kvAddr, "kv-addr", ":50051", "Client-facing KV service address")
	flag.StringVar(&raftAddr, "raft-addr", ":60051", "Internal Raft communication address")
	flag.Var(&peers, "peer", "Peer address (provide multiple times)")
	flag.Parse()

	if id == "" { log.Fatal("-id flag is required") }

	peerMap := make(map[string]string)
	for _, p := range peers {
		parts := strings.Split(p, "=")
		if len(parts) != 2 { log.Fatalf("Invalid peer format: %s. Expected id=address", p) }
		peerMap[parts[0]] = parts[1]
	}
	log.Printf("Starting node %s. Peers: %v", id, peerMap)

	// --- NEW STARTUP ORDER ---

	// 1. Initialize modules (but don't connect or start them yet)
	raftNode := raft.NewNode(id)
	
	kvStore, err := store.NewLevelDBStore("./data/" + id)
	if err != nil { log.Fatalf("Failed to create store: %v", err) }
	defer kvStore.Close()
	kvServiceServer := rpc.NewKVServiceServer(kvStore)
	
	// 2. Start the network servers in the background
	go startRaftServer(raftNode, raftAddr)
	go startKVServer(kvServiceServer, kvAddr)

	// 3. Wait a moment for servers to come up. This is a simple solution for local testing.
	time.Sleep(1 * time.Second)

	// 4. Now, connect the Raft node to its peers.
	raftNode.ConnectToPeers(peerMap)

	// 5. Finally, start the Raft node's main logic.
	raftNode.Start()

	// --- Graceful Shutdown ---
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh
	
	raftNode.Stop()
	log.Println("Shutdown complete.")
}

func startRaftServer(raftNode *raft.Node, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil { log.Fatalf("Failed to listen on raft-addr %s: %v", addr, err) }

	server := grpc.NewServer()
	pb.RegisterRaftInternalServer(server, raftNode)
	log.Printf("Raft server listening on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve Raft gRPC: %v", err)
	}
}

func startKVServer(kvService *rpc.Server, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil { log.Fatalf("Failed to listen on kv-addr %s: %v", addr, err) }

	server := grpc.NewServer()
	pb.RegisterKVStoreServer(server, kvService)
	log.Printf("KV server listening on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve KV gRPC: %v", err)
	}
}