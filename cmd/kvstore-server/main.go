package main

import (
	"fmt"
	"log"

	"Distributed_KV_store/internal/store" // Import our new store package
)

func main() {
	log.Println("Starting Standalone Key-Value Store...")

	// Define the path for our database files.
	// This will create a 'data' directory in your project root.
	dbPath := "./data"

	// Create a new store instance.
	kvStore, err := store.NewLevelDBStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	// `defer` ensures that the store's Close() method is called right before
	// the main function exits, even if an error occurs.
	defer kvStore.Close()

	log.Println("Store created successfully.")

	// --- Let's test the store ---

	key := "hello"
	value := "world"

	// 1. Put a key-value pair
	log.Printf("Putting key='%s', value='%s'", key, value)
	if err := kvStore.Put(key, value); err != nil {
		log.Fatalf("Failed to put: %v", err)
	}
	log.Println("Put successful.")

	// 2. Get the key back
	log.Printf("Getting key='%s'", key)
	retrievedValue, err := kvStore.Get(key)
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}
	log.Printf("Get successful. Retrieved value: '%s'", retrievedValue)

	// Verify the value
	if retrievedValue != value {
		log.Fatalf("Value mismatch! Expected '%s', got '%s'", value, retrievedValue)
	}

	// 3. Delete the key
	log.Printf("Deleting key='%s'", key)
	if err := kvStore.Delete(key); err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}
	log.Println("Delete successful.")

	// 4. Try to get the deleted key (expect an error)
	log.Printf("Getting deleted key='%s' (expecting an error)", key)
	_, err = kvStore.Get(key)
	if err != nil {
		log.Printf("Successfully failed to get deleted key, as expected. Error: %v", err)
	} else {
		log.Fatalf("Should have failed to get deleted key, but it succeeded.")
	}

	fmt.Println("\n✅ All operations completed successfully!")
}