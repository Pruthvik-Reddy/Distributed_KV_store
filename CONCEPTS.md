Go : 
    - Install Go and check version with "go version". 
    
LevelDB :   
    - LevelDB is a fast, lightweight, key-value storage library written by Google.
    - Stored data as key-value pairs. 
    - Keys are sorted in lexicographical sorted order. 

    - 
    - Download the levelDB library using "go get github.com/syndtr/goleveldb/leveldb". It will add the 
      LevelDB dependency to the go.mod and go.sum files. 

gRPC : 
    - Open source Remote Procedure Call developed byb Google. 
    - Allows services to communicate across different services, machines, etc in a microservices architecture. 

    In microservices, it is harder for communication across different services. 
    - gRPC enables faster communication between services compared to REST+JSON and a big part of that speed comes
      from using protocol buffers ( protobuf)
    - .proto files define how communication happens and it occurs over HTTP ( HTTP + .proto)
    - It defines data types, contracts, deadlines, etc and can also enable communication across different
    services in different languages. 
    - It enables communication between services as they were local even if they are on different machines. 
    - protobuf is faster compared to JSON and XML and has cross-language support. 

    proto file : 
        - defines API contract. 
        
RAFT : 
    How RAFT works : 
    1. Simulate a cluster by starting different servers on different ports. 
    2. Each server starts as a follower. 
    3. After a random timeout, if a follower does not hear back from leader, the follower declares itself as candidate and start election. 
    4. It requests votes from all other nodes. 
    5. If it receives votes from a majority of the nodes, it will become the Leader.
    6. The new Leader will then periodically send "heartbeat" messages to all Followers to assert its authority and prevent new elections.

    Strong Consistency, Writes and Reads - How they work
    1. When the client sends a PUT / DELETE request to any node, it should be redirected to the leader who processes that request. ( the follower nodes should forward it to leader)

    2. The leader then should make an entry of it in log file. From the RAFT log file, the other followers implement it. 

    3. We are focusing on Strong consistency here, so initially the reads are served by leader only. 

    4. The leader responds to the client only after the operation has been committed by Raft and applied to LevelDB database. 


    Example : 
    1. If Node 2 isn't the leader, it will reject the request with a "not the leader" error.

    2.If Node 2 is the leader, it will not immediately write to its own database. Instead, it will create a "log entry" representing the PUT command.

    3. It will send this new log entry to all its followers via an AppendEntries RPC.

    4.When a majority of followers have successfully saved the entry to their own logs, the leader considers the entry "committed"

    5.The leader then "applies" the command from the log entry to its own LevelDB store.

    6.At the same time, the followers also apply the committed command to their local LevelDB stores.

    7.Only after the leader has applied the change does it respond with "success" to the client.


Challenges : 
    1. Not a significant challenge but had to get used to Go to start coding. 
    2. Writing the protobuf files. They were little complex to understand. 
    3. RACE conditions while implementing RAFT : 
        -  Looking at the logs, especially from Node 2 and 3:

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

    4. Lots of deadlocks. 

Decisions : 
