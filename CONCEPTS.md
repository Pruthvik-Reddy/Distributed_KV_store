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
        
    

Challenges : 
    1. Not a significant challenge but had to get used to Go to start coding. 
    2. Writing the protobuf files. They were little complex to understand. 
    3. 

Decisions : 
