// File: internal/raft/raft.go
package raft

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "distributed-kv-store/api/kvstore/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

const electionTimeoutMin = 250 * time.Millisecond
const heartbeatInterval = 100 * time.Millisecond

type Node struct {
	pb.UnimplementedRaftInternalServer

	mu        sync.RWMutex
	id        string
	state     NodeState
	peers     map[string]pb.RaftInternalClient
	shutdown  chan struct{}

	currentTerm uint64
	votedFor    string

	electionTimer *time.Timer
}

// NewNode now only initializes the struct, it does not connect to peers.
func NewNode(id string) *Node {
	node := &Node{
		id:       id,
		state:    Follower,
		peers:    make(map[string]pb.RaftInternalClient),
		shutdown: make(chan struct{}),
		votedFor: "",
	}
	return node
}

// ConnectToPeers establishes gRPC connections to all peers.
// This should be called after the node's own gRPC server has started.
func (n *Node) ConnectToPeers(peerAddrs map[string]string) {
	for peerId, addr := range peerAddrs {
		if peerId == n.id {
			continue
		}
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("[%s] Failed to connect to peer %s: %v", n.id, peerId, err)
		}
		n.peers[peerId] = pb.NewRaftInternalClient(conn)
		log.Printf("[%s] Connected to peer %s at %s", n.id, peerId, addr)
	}
}

func (n *Node) Start() {
	n.mu.Lock()
	n.resetElectionTimer()
	n.mu.Unlock()

	go func() {
		for {
			select {
			case <-n.shutdown:
				log.Printf("[%s] Shutting down", n.id)
				return
			case <-n.electionTimer.C:
				n.mu.RLock()
				log.Printf("[%s] Election timer fired (state: %v)", n.id, n.state)
				n.mu.RUnlock()
				n.startElection()
			}
		}
	}()
}

func (n *Node) Stop() {
	close(n.shutdown)
}

func (n *Node) resetElectionTimer() {
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	timeout := electionTimeoutMin + time.Duration(rand.Intn(150))*time.Millisecond
	n.electionTimer = time.NewTimer(timeout)
}

func (n *Node) convertToFollower(term uint64) {
	// Must be called with lock held
	n.state = Follower
	n.currentTerm = term
	n.votedFor = ""
	log.Printf("[%s] became Follower at term %d", n.id, n.currentTerm)
	n.resetElectionTimer()
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.id
	log.Printf("[%s] became Candidate, starting election for term %d", n.id, n.currentTerm)
	n.resetElectionTimer()

	term := n.currentTerm
	candidateId := n.id
	n.mu.Unlock()

	votes := 1

	var wg sync.WaitGroup
	for peerId := range n.peers {
		wg.Add(1)
		go func(peerId string) {
			defer wg.Done()
			args := &pb.RequestVoteArgs{
				Term:        term,
				CandidateId: candidateId,
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			reply, err := n.peers[peerId].RequestVote(ctx, args)
			if err != nil {
				// We don't log this anymore to reduce noise, a timeout is sufficient
				// log.Printf("[%s] failed to send RequestVote to %s: %v", n.id, peerId, err)
				return
			}
			
			n.mu.Lock()
			defer n.mu.Unlock()
			
			// If we are no longer a candidate, ignore the reply
			if n.state != Candidate {
				return
			}
			
			if reply.Term > n.currentTerm {
				n.convertToFollower(reply.Term)
				return
			}

			if reply.VoteGranted {
				votes++
				if votes > (len(n.peers)+1)/2 {
					n.convertToLeader()
				}
			}
		}(peerId)
	}
}

func (n *Node) convertToLeader() {
	// Must be called with lock held
	if n.state != Candidate {
		return
	}
	n.state = Leader
	log.Printf("**************************************************")
	log.Printf("[%s] became LEADER for term %d", n.id, n.currentTerm)
	log.Printf("**************************************************")
	
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}

	go n.sendHeartbeats()
}

func (n *Node) sendHeartbeats() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.shutdown:
			return
		case <-ticker.C:
			n.mu.RLock()
			if n.state != Leader {
				n.mu.RUnlock()
				return
			}
			// In the next task, this is where we'll send real AppendEntries RPCs.
			// log.Printf("[%s] (Leader) sending heartbeats", n.id)
			n.mu.RUnlock()
		}
	}
}

func (n *Node) RequestVote(ctx context.Context, args *pb.RequestVoteArgs) (*pb.RequestVoteReply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.currentTerm {
		return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: false}, nil
	}

	if args.Term > n.currentTerm {
		n.convertToFollower(args.Term)
	}

	if n.votedFor == "" || n.votedFor == args.CandidateId {
		n.votedFor = args.CandidateId
		n.resetElectionTimer()
		log.Printf("[%s] Voted FOR %s at term %d", n.id, args.CandidateId, n.currentTerm)
		return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: true}, nil
	}
	
	return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: false}, nil
}