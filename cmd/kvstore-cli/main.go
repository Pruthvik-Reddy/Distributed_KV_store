package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "distributed-kv-store/api/kvstore/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// We expect the command as the first argument, e.g., "get", "put", "delete".
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args...]")
		os.Exit(1)
	}

	// Server address to connect to.
	serverAddr := "localhost:50051"

	// Set up a connection to the server.
	// We use `WithTransportCredentials(insecure.NewCredentials())` because we are not using TLS for this example.
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// Create a client stub from the connection.
	client := pb.NewKVStoreClient(conn)

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Use a switch to decide which RPC to call based on the command-line argument.
	command := os.Args[1]
	switch command {
	case "get":
		if len(os.Args) != 3 {
			log.Fatalf("Usage: client get <key>")
		}
		key := os.Args[2]
		res, err := client.Get(ctx, &pb.GetRequest{Key: key})
		if err != nil {
			log.Fatalf("Could not get value for key %s: %v", key, err)
		}
		log.Printf("GET Response: Value = '%s'", res.GetValue())

	case "put":
		if len(os.Args) != 4 {
			log.Fatalf("Usage: client put <key> <value>")
		}
		key := os.Args[2]
		value := os.Args[3]
		_, err := client.Put(ctx, &pb.PutRequest{Key: key, Value: value})
		if err != nil {
			log.Fatalf("Could not put key %s: %v", key, err)
		}
		log.Printf("PUT successful for key '%s'", key)

	case "delete":
		if len(os.Args) != 3 {
			log.Fatalf("Usage: client delete <key>")
		}
		key := os.Args[2]
		_, err := client.Delete(ctx, &pb.DeleteRequest{Key: key})
		if err != nil {
			log.Fatalf("Could not delete key %s: %v", key, err)
		}
		log.Printf("DELETE successful for key '%s'", key)

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}