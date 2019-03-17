package server

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/client"
	"github.com/IvanProdaiko94/raft-protocol-implementation/consensus"
	"github.com/IvanProdaiko94/raft-protocol-implementation/log"
	schemapb "github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"github.com/dvln/out"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	Follower = iota
	Candidate
	Leader
)

var min = 150
var max = 300

func randomTime(min, max int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Millisecond * time.Duration(min+rand.Intn(max-min+1))
}

func randomTimeout(min, max int) {
	time.Sleep(randomTime(min, max))
}

type Node struct {
	term      *int32
	state     *int32
	votedFor  uuid.UUID
	id        uuid.UUID
	log       *log.Log
	heartbeat chan time.Time
	clients   []schemapb.NodeClient
	server    *grpc.Server
}

/**
While waiting for votes, a candidate may receive an AppendEntries RPC from another server claiming to be leader.
If the leader’s term (included in its RPC) is at least as large as the candidate’s current term, then the candidate
recognizes the leader as legitimate and returns to follower state.
If the term in the RPC is smaller than the candidate’s current term,
then the candidate rejects the RPC and continues in candidate state.
*/
func (n *Node) AppendEntries(ctx context.Context, input *schemapb.AppendEntriesInput) (*schemapb.AppendEntriesOutput, error) {
	n.heartbeat <- time.Now()
	// Reply false if term < currentTerm
	if atomic.LoadInt32(n.term) < input.Term {
		return &schemapb.AppendEntriesOutput{
			Term:    atomic.LoadInt32(n.term),
			Success: false,
		}, nil
	}
	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	// FIXME
	lastLogEntry := n.log.EntryByIndex(int(input.PrevLogIndex))
	if lastLogEntry != nil || lastLogEntry.Term != int(input.PrevLogTerm) {
		return &schemapb.AppendEntriesOutput{
			Term:    atomic.LoadInt32(n.term),
			Success: false,
		}, nil
	}
	// If an existing entry conflicts with a new one (same index but different terms),
	// delete the existing entry and all that follow it
	// TODO

	// Append any new entries not already in the log
	var entries = make([]log.Entry, len(input.Entries))
	for i, e := range input.Entries {
		entry, err := log.EntryFromBytes(e.GetValue(), int(atomic.LoadInt32(n.term)))
		if err != nil {
			return nil, err
		}
		entries[i] = *entry
	}
	n.log.AppendEntries(entries)
	// If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	// TODO
	return nil, nil
}

/**
The RequestVote RPC implements this restriction: the RPC includes information about the candidate’s log, and the
voter denies its vote if its own log is more up-to-date than that of the candidate.

Raft determines which of two logs is more up-to-date by comparing the index and term of the last entries in the
logs. If the logs have last entries with different terms, then the log with the later term is more up-to-date. If the logs
end with the same term, then whichever log is longer is more up-to-date.
*/
func (n *Node) RequestVote(ctx context.Context, input *schemapb.RequestVoteInput) (*schemapb.RequestVoteOutput, error) {
	//if n.log.LastTerm >= int(input.LastLogTerm) && n.log.LastIndex > int(input.LastLogIndex) {
	//	votedFor, err := uuid.FromBytes([]byte(input.CandidateId.GetValue()))
	//	if err != nil {
	//		return nil, err
	//	}
	//	n.RunAsFollower(ctx)
	//	n.term = int(input.Term)
	//	n.votedFor = votedFor
	//	return &schemapb.RequestVoteOutput{
	//		Term: int64(n.term),
	//		VoteGranted:true,
	//	}, nil
	//}
	return &schemapb.RequestVoteOutput{
		Term:        atomic.LoadInt32(n.term),
		VoteGranted: false,
	}, nil
}

/**
When servers start up, they begin as followers. A server remains in follower state as long as it receives valid
RPCs from a leader or candidate. Leaders send periodic heartbeats (AppendEntries RPCs that carry no log entries)
to all followers in order to maintain their authority. If a follower receives no communication over a period of time
called the election timeout, then it assumes there is no viable leader and begins an election to choose a new leader.
*/
func (n *Node) RunAsFollower(ctx context.Context) {
	out.Infof("Change role to Follower. Term: %d\n", n.term)
	atomic.StoreInt32(n.state, Follower)
	go func() {
		timer := time.NewTimer(randomTime(1500, 2000))
		for {
			// do not check heartbeats or election timeout.
			// If there appears another leader it will force to RunAsFollower once again.
			if atomic.LoadInt32(n.state) == Leader {
				out.Infoln("Timer stopped")
				timer.Stop()
				return
			}
			select {
			case <-timer.C:
				timeout := randomTime(1500, 2000)
				cctx, cancel := context.WithTimeout(ctx, timeout)
				timer.Reset(timeout)
				n.RunAsCandidate(cctx, cancel)
			case <-n.heartbeat:
				timer.Reset(randomTime(1500, 2000))
			}
		}
	}()
}

/**
To begin an election, a follower increments its current term and transitions to candidate state. It then votes for
itself and issues RequestVote RPCs in parallel to each of the other servers in the cluster.
A candidate continues in this state until one of three things happens:
	- (a) it wins the election
	- (b) another server establishes itself as leader
	- (c) a period of time goes by with no winner
*/
func (n *Node) RunAsCandidate(ctx context.Context, cancelFunc context.CancelFunc) {
	atomic.AddInt32(n.term, 1)
	out.Infof("Change role to Candidate. Term: %d\n", n.term)
	atomic.StoreInt32(n.state, Candidate)
	var count *int32
	for _, c := range n.clients {
		go func() {
			output, err := c.RequestVote(ctx, &schemapb.RequestVoteInput{
				Term:        atomic.LoadInt32(n.term),
				CandidateId: &wrappers.StringValue{Value: n.id.String()},
				//LastLogIndex:
				//LastLogTerm:
			})
			if err != nil {
				out.Error(err)
			}
			if output.VoteGranted {
				atomic.AddInt32(count, 1)
			}
			if consensus.Reached(int(atomic.LoadInt32(count)), len(n.clients)) {
				// it wins the election
				n.RunAsLeader(context.Background())
				cancelFunc()
			}
		}()
	}
}

/**
Send heartbeat to all of the followers
*/
func (n *Node) RunAsLeader(ctx context.Context) {
	out.Infof("Change role to Leader. Term: %d\n", n.term)
	atomic.StoreInt32(n.state, Leader)
	timer := time.NewTimer(time.Millisecond * randomTime(min, max))
	go func() {
		for {
			<-timer.C
			for _, c := range n.clients {
				_, err := c.AppendEntries(ctx, &schemapb.AppendEntriesInput{
					Term:     atomic.LoadInt32(n.term),
					LeaderId: &wrappers.StringValue{Value: n.id.String()},
					// TODO Change these 2 fields
					//PrevLogIndex: nil,
					//PrevLogTerm:  nil,
				})
				if err != nil {
					out.Error(err)
				}
			}
		}
	}()
}

func (n *Node) Launch(addr string) error {
	if err := Listen(n.server, addr); err != nil {
		return err
	}
	return nil
}

func (n *Node) Stop() {
	n.server.GracefulStop()
}

func (n *Node) InitClients(addressList []string) (err error) {
	n.clients, err = client.CreateMultiple(addressList)
	if err != nil {
		return err
	}
	return nil
}

func New() *Node {
	state := int32(Follower)
	term := int32(0)
	n := &Node{
		state:     &state,
		term:      &term,
		id:        uuid.New(),
		heartbeat: make(chan time.Time),
		log:       log.New(),
	}
	n.server = CreateGRPC(n)
	return n
}
