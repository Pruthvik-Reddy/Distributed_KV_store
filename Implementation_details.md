
Installations : 
1. Go (go version)
2. LevelDB : 
        - It is a C++ library and needs a high-level binding ( like Go or Node) to avoid manually compiling. 
        
        - Download the levelDB library using "go get github.com/syndtr/goleveldb/leveldb". It will add the 
          LevelDB dependency to the go.mod and go.sum files. 


Step 1 : Create initial structure for the project. 

Step 2 : Creating basic Put/Get/Delete operations using levelDB without any networking. 
    a. go get github.com/syndtr/goleveldb/leveldb
    b. Implementing store module (internal/store/store.go) 
        - We need to implement an interface with get, put and delete operations. 

    c. Create a temporary main file in cmd/kvstore-server/main.go for temporarily creating levelDB instance and performing put,delete,get operations. 

        - Create a folder ./data if it does not exist and check if any levelDB instance exists. If not create one. Else use the existing one. 
        - We can only create one LevelDB instance inside that ./data folder as LevelDB is designed to store only one logical database per directory. 

    Step 2 Successfully Completed. 

Step 3 : Adding Network Layer using gRPC. 
    - gRPC communication. 

    - Until step 2, the key-value store runs only on one machine and used only by one current running program. 
    - I want it to now evolve into a client-server application, where client can connect to server and do get/put/delete 
    operations. 
    
    1. Define API contract : define put, get and delete operations in protobuf files. 
    2. Build Server : kvstore-server - It will start as a gRPC server, listen for requests on a port and process them 
                        in store module. 
    3. Build Client : kvstore-cli, connects to server and sends requests and then should print response from server. 

    - Install protocol buffer compiler (protoc) to generate code. 

Step 4 : Adding RAFT for leader election. 