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
import "distributed-kv-store/internal/monitoring"


type stringslice []string
// ... (stringslice functions remain unchanged)
func (s *stringslice) String() string { return strings.Join(*s, ", ") }
func (s *stringslice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var id, kvAddr, raftAddr, metricsAddr string
	var peers stringslice
	flag.StringVar(&id, "id", "", "Node ID (required)")
	flag.StringVar(&kvAddr, "kv-addr", ":50051", "Client-facing KV service address")
	flag.StringVar(&raftAddr, "raft-addr", ":60051", "Internal Raft communication address")
	flag.StringVar(&metricsAddr, "metrics-addr", ":9091", "Metrics server address")
	flag.Var(&peers, "peer", "Peer address (provide multiple times)")
	flag.Parse()

	if id == "" { log.Fatal("-id flag is required") }

	peerMap := make(map[string]string)
	for _, p := range peers {
		parts := strings.Split(p, "=");
		if len(parts) != 2 { log.Fatalf("Invalid peer format: %s. Expected id=address", p) }
		peerMap[parts[0]] = parts[1]
	}
	log.Printf("Starting node %s. Peers: %v", id, peerMap)

	// --- NEW: Create the apply channel ---
	applyCh := make(chan raft.ApplyMsg)

	// 1. Initialize modules
	// NewNode now requires the apply channel
	raftNode := raft.NewNode(id, applyCh)
	
	kvStore, err := store.NewLevelDBStore("./data/" + id)
	if err != nil { log.Fatalf("Failed to create store: %v", err) }
	defer kvStore.Close()

	// NewKVServiceServer now requires the Raft node
	kvServiceServer := rpc.NewKVServiceServer(kvStore, raftNode)
	
	// --- NEW: Start the command application goroutine ---
	go applyCommands(kvStore, applyCh)

	// 2. Start servers
	monitoring.StartMetricsServer(metricsAddr)
	go startRaftServer(raftNode, raftAddr)
	go startKVServer(kvServiceServer, kvAddr)

	time.Sleep(1 * time.Second)

	// 3. Connect to peers
	raftNode.ConnectToPeers(peerMap)

	// 4. Start Raft
	raftNode.Start()

	// --- Graceful Shutdown ---
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh
	
	raftNode.Stop()
	log.Println("Shutdown complete.")
}

// applyCommands is a long-running goroutine that applies committed commands to the KV store.
func applyCommands(s store.Store, applyCh <-chan raft.ApplyMsg) {
	log.Println("Started command application loop")
	for msg := range applyCh {
		if msg.CommandValid {
			// Simple command parsing: "OP:KEY:VALUE"
			parts := strings.SplitN(string(msg.Command), ":", 3)
			op := parts[0]
			key := parts[1]

			switch op {
			case "PUT":
				value := parts[2]
				if err := s.Put(key, value); err != nil {
					log.Printf("Failed to apply PUT command: %v", err)
				}
			case "DEL":
				if err := s.Delete(key); err != nil {
					log.Printf("Failed to apply DEL command: %v", err)
				}
			}
		}
	}
}


func startRaftServer(raftNode *raft.Node, addr string) { /* ... unchanged ... */
	lis, err := net.Listen("tcp", addr)
	if err != nil { log.Fatalf("Failed to listen on raft-addr %s: %v", addr, err) }
	server := grpc.NewServer()
	pb.RegisterRaftInternalServer(server, raftNode)
	log.Printf("Raft server listening on %s", addr)
	if err := server.Serve(lis); err != nil { log.Fatalf("Failed to serve Raft gRPC: %v", err) }
}
func startKVServer(kvService *rpc.Server, addr string) { /* ... unchanged ... */
	lis, err := net.Listen("tcp", addr)
	if err != nil { log.Fatalf("Failed to listen on kv-addr %s: %v", addr, err) }
	server := grpc.NewServer()
	pb.RegisterKVStoreServer(server, kvService)
	log.Printf("KV server listening on %s", addr)
	if err := server.Serve(lis); err != nil { log.Fatalf("Failed to serve KV gRPC: %v", err) }
}