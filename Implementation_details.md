
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
    - Need to create nodes that communicate with each other, elect a leader and then check if nodes are active through heartbeats. 

    - For this step, I am not yet connecting with the key-value store. I just want to ensure if my RAFT algorithm works. 

    1. Simulate a cluster by starting different servers on different ports. 
    2. Each server starts as a follower. 
    3. After a random timeout, if a follower does not hear back from leader, the follower declares itself as candidate and start election. 
    4. It requests votes from all other nodes. 
    5. If it receives votes from a majority of the nodes, it will become the Leader.
    6. The new Leader will then periodically send "heartbeat" messages to all Followers to assert its authority and prevent new elections.

    Implementation : 
    1. Define new api contract for RAFT internal communication. This should only be for RAFT ( server - server communication). 
    2. Define a RAFT Node. Use struct from golang to define Node which contains all the Raft state (term, state, timers) and the logic for state transitions (e.g., convertToCandidate, startElection). It should also define the logic whether to grant a vote for candidate or not. 
    3. Need to integrate RAFT into main ( this is different from simple starting a server ) : 
        - Parse more command-line flags to know its own ID, its own address, and the addresses of its peers.

        - Start two gRPC servers: one for client-facing KV requests (on one port) and another for internal Raft communication (on a different port).

    
    - Open three seperate terminals and start three different servers and see the logs to check if the code is working as desired. 

    Challenges : RACE conditions while implementing RAFT.
        Looking at the logs, especially from Node 2 and 3:

        connection ... actively refused it: This is the key. It means a node tried to make a gRPC call to a peer, but the peer's gRPC server wasn't listening for connections yet.

        Term numbers skyrocketing (term 64, 65, ...): This happens when nodes can't communicate. They all time out, become candidates, fail to get votes (because the RPCs fail), and then time out again, starting a new election with a higher term.

        The root cause is in main.go. You call raft.NewNode(), which immediately tries to connect to its peers. However, you don't start the gRPC servers (startRaftServer) until after NewNode() is called. The nodes are trying to connect to each other before any of them are actually listening.

        The Fix
        We need to change the order of operations:

        Initialize the Raft Node struct (without connecting).

        Start the gRPC servers so they are listening.

        Add a small delay to ensure they are ready.

        Then, tell the Raft node to connect to its peers.

        Finally, start the Raft node's main logic (the election timers). 


Step 5 : Integrating RAFT with Key-Value Store 

    When the client sends a PUT / DELETE request to any node, it should be redirected to the leader who processes that request. ( the follower nodes should forward it to leader)

    The leader then should make an entry of it in log file. From the RAFT log file, the other followers implement it. 

    We are focusing on Strong consistency here, so initially the reads are served by leader only. 

    The leader responds to the client only after the operation has been committed by Raft and applied to LevelDB database. 


Running : 
    - Initialize 3 nodes as RAFT nodes and then use another CLI to run as client and make requests. 

    If you make a put request, I should be able to see the logs in CLI. 


Step 6 : Adding Observability, Metrics using Grafana and Prometheus and logging with ELK. 

    - I do not want anything complex as of now. I just want to see a dashoboard on how the leader is, how many PUT or GET requests are being handled and how up-to date are the logs. 
    Prometheus is the time-series database to store these and Grafana can be used for dashboards. 

    1. Add prometheus dependency. 
    2. create a http based /metrics endpoint on a different port to expose the current values to prometheus. 
    3. Make changes to codebase ( put, get methods increment counter, updating leader) to call the monitoring functions at right moments. 
    4. Use Docker and Docker compose to run them locally in a easy way. 


