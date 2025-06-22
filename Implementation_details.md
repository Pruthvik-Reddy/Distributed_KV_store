
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

Step 4 : Adding RAFT for leader election. 