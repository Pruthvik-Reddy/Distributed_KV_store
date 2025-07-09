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

// ApplyMsg is sent on the applyCh channel when a command has been committed.
type ApplyMsg struct {
	CommandValid bool
	Command      []byte
	CommandIndex uint64
}

type Node struct {
	pb.UnimplementedRaftInternalServer

	mu        sync.RWMutex
	id        string
	state     NodeState
	peers     map[string]pb.RaftInternalClient
	applyCh   chan ApplyMsg
	shutdown  chan struct{}

	// Persistent state on all servers
	currentTerm uint64
	votedFor    string
	log         []*pb.LogEntry

	// Volatile state on all servers
	commitIndex uint64
	lastApplied uint64

	// Volatile state on leaders
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	electionTimer *time.Timer
}

func NewNode(id string, applyCh chan ApplyMsg) *Node {
	node := &Node{
		id:          id,
		state:       Follower,
		peers:       make(map[string]pb.RaftInternalClient),
		applyCh:     applyCh,
		shutdown:    make(chan struct{}),
		votedFor:    "",
		log:         []*pb.LogEntry{{Term: 0}}, // Dummy entry at index 0
		commitIndex: 0,
		lastApplied: 0,
	}
	return node
}

// Propose is called by the application layer (KV service) to submit a new command.
// It will only succeed if the node is the leader.
// THIS IS THE METHOD THAT WAS MISSING.
func (n *Node) Propose(command []byte) (index uint64, term uint64, isLeader bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != Leader {
		return 0, 0, false
	}

	term = n.currentTerm
	entry := &pb.LogEntry{
		Term:    term,
		Command: command,
	}
	n.log = append(n.log, entry)
	index = uint64(len(n.log) - 1)
	log.Printf("[%s] (Leader) Proposed new log entry at index %d, term %d", n.id, index, term)

	// We don't wait for replication here. The leader's main loop will handle it.
	return index, term, true
}


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
				n.startElection()
			}
		}
	}()
}

func (n *Node) Stop() {
	close(n.shutdown)
}

// This function assumes the caller is holding the lock.
func (n *Node) convertToFollower(term uint64) {
	n.state = Follower
	n.currentTerm = term
	n.votedFor = ""
	log.Printf("[%s] became Follower at term %d", n.id, n.currentTerm)
	n.resetElectionTimer()
}

func (n *Node) resetElectionTimer() {
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	timeout := electionTimeoutMin + time.Duration(rand.Intn(150))*time.Millisecond
	n.electionTimer = time.NewTimer(timeout)
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.id
	log.Printf("[%s] became Candidate, starting election for term %d", n.id, n.currentTerm)
	n.resetElectionTimer()

	lastLogIndex := uint64(len(n.log) - 1)
	lastLogTerm := n.log[lastLogIndex].Term
	term := n.currentTerm
	candidateId := n.id
	n.mu.Unlock() // Release lock before sending RPCs

	votes := 1

	for peerId := range n.peers {
		go func(peerId string) {
			args := &pb.RequestVoteArgs{
				Term:         term,
				CandidateId:  candidateId,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			reply, err := n.peers[peerId].RequestVote(ctx, args)
			if err != nil {
				return
			}
			
			n.mu.Lock()
			defer n.mu.Unlock()
			
			if n.state != Candidate || n.currentTerm != term {
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

	n.nextIndex = make(map[string]uint64)
	n.matchIndex = make(map[string]uint64)
	lastLogIndex := uint64(len(n.log) - 1)
	for peerId := range n.peers {
		n.nextIndex[peerId] = lastLogIndex + 1
		n.matchIndex[peerId] = 0
	}
	
	log.Printf("**************************************************")
	log.Printf("[%s] became LEADER for term %d", n.id, n.currentTerm)
	log.Printf("**************************************************")
	
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	go n.sendReplication()
}

func (n *Node) sendReplication() {
	for {
		n.mu.RLock()
		if n.state != Leader {
			n.mu.RUnlock()
			return
		}
		n.mu.RUnlock()

		for peerId := range n.peers {
			go n.replicateToPeer(peerId)
		}
		
		time.Sleep(heartbeatInterval)
	}
}

func (n *Node) replicateToPeer(peerId string) {
	n.mu.RLock()
	if n.state != Leader {
		n.mu.RUnlock()
		return
	}

	prevLogIndex := n.nextIndex[peerId] - 1
	if prevLogIndex > uint64(len(n.log)-1) {
		prevLogIndex = uint64(len(n.log)-1)
	}
	prevLogTerm := n.log[prevLogIndex].Term
	
	var entries []*pb.LogEntry
	if uint64(len(n.log)-1) >= n.nextIndex[peerId] {
		entries = n.log[n.nextIndex[peerId]:]
	}

	args := &pb.AppendEntriesArgs{
		Term:         n.currentTerm,
		LeaderId:     n.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: n.commitIndex,
	}
	n.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	reply, err := n.peers[peerId].AppendEntries(ctx, args)
	if err != nil { return }

	n.mu.Lock()
	defer n.mu.Unlock()

	if reply.Term > n.currentTerm {
		n.convertToFollower(reply.Term)
		return
	}

	if n.state == Leader && reply.Term == n.currentTerm {
		if reply.Success {
			n.matchIndex[peerId] = args.PrevLogIndex + uint64(len(args.Entries))
			n.nextIndex[peerId] = n.matchIndex[peerId] + 1
			n.updateCommitIndex()
		} else {
			if n.nextIndex[peerId] > 1 {
				n.nextIndex[peerId]--
			}
		}
	}
}

func (n *Node) updateCommitIndex() {
	// Must be called with lock held
	for N := uint64(len(n.log) - 1); N > n.commitIndex && n.log[N].Term == n.currentTerm; N-- {
		count := 1
		for peerId := range n.peers {
			if n.matchIndex[peerId] >= N {
				count++
			}
		}

		if count > (len(n.peers)+1)/2 {
			n.commitIndex = N
			log.Printf("[%s] (Leader) Set commitIndex to %d", n.id, n.commitIndex)
			n.applyCommitted()
			break
		}
	}
}

func (n *Node) applyCommitted() {
	// Must be called with lock held
	for n.commitIndex > n.lastApplied {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		log.Printf("[%s] Applying command at index %d", n.id, n.lastApplied)
		n.applyCh <- ApplyMsg{
			CommandValid: true,
			Command:      entry.Command,
			CommandIndex: n.lastApplied,
		}
	}
}

// ---- RPC Handlers ----

func (n *Node) RequestVote(ctx context.Context, args *pb.RequestVoteArgs) (*pb.RequestVoteReply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.currentTerm {
		return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: false}, nil
	}
	if args.Term > n.currentTerm {
		n.convertToFollower(args.Term)
	}

	lastLogIndex := uint64(len(n.log) - 1)
	lastLogTerm := n.log[lastLogIndex].Term
	uptodate := args.LastLogTerm > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)

	if (n.votedFor == "" || n.votedFor == args.CandidateId) && uptodate {
		n.votedFor = args.CandidateId
		n.resetElectionTimer()
		log.Printf("[%s] Voted FOR %s at term %d", n.id, args.CandidateId, n.currentTerm)
		return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: true}, nil
	}
	
	return &pb.RequestVoteReply{Term: n.currentTerm, VoteGranted: false}, nil
}

func (n *Node) AppendEntries(ctx context.Context, args *pb.AppendEntriesArgs) (*pb.AppendEntriesReply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.currentTerm {
		return &pb.AppendEntriesReply{Term: n.currentTerm, Success: false}, nil
	}
	
	n.resetElectionTimer()

	if args.Term > n.currentTerm {
		n.convertToFollower(args.Term)
	}

	if args.PrevLogIndex >= uint64(len(n.log)) || n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return &pb.AppendEntriesReply{Term: n.currentTerm, Success: false}, nil
	}

	if len(args.Entries) > 0 {
		n.log = n.log[:args.PrevLogIndex+1]
		n.log = append(n.log, args.Entries...)
		log.Printf("[%s] (Follower) Appended %d entries. New log length: %d", n.id, len(n.log))
	}

	if args.LeaderCommit > n.commitIndex {
		n.commitIndex = min(args.LeaderCommit, uint64(len(n.log)-1))
		n.applyCommitted()
	}
	
	return &pb.AppendEntriesReply{Term: n.currentTerm, Success: true}, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}