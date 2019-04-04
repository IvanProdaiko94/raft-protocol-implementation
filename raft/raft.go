package raft

import (
	"context"
	"fmt"
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	"github.com/IvanProdaiko94/raft-protocol-implementation/rpc"
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"time"
)

const (
	requestTimeout  = time.Millisecond * 100
	electionTimeout = time.Millisecond * 1000
)

type Raft interface {
	Start(ctx context.Context)
	Stop()
	Stats() map[string]string
	GetState() *State
}

type raft struct {
	state       *State
	logger      *Logger
	log         *Log
	server      rpc.Server
	clients     []rpc.Client
	clusterSize int
	addresses   []env.Node
	leaderId    int32 // required for client redirect
	config      env.Node
}

func New(config env.Node, nodesAddresses []env.Node, clusterSize int) Raft {
	return &raft{
		state:       NewState(),
		logger:      NewLogger(),
		server:      rpc.NewServer(config),
		clients:     rpc.NewClients(nodesAddresses),
		clusterSize: clusterSize,
		addresses:   nodesAddresses,
		leaderId:    -1,
		config:      config,
		log:         NewLog(),
	}
}

func (r *raft) GetState() *State {
	return r.state
}

func (r *raft) Stop() {
	r.server.Stop()
}

func (r *raft) Start(ctx context.Context) {
	r.state.SetState(Follower)

	go func() { // start server
		r.logger.Log(
			PrefixInfo,
			fmt.Sprintf("Started grpc server on address: %s", r.config.Address),
		)
		if err := r.server.Start(ctx); err != nil {
			panic(err)
		}
	}()

	for _, client := range r.clients { // init clients
		if err := client.Deal(); err != nil {
			panic(err)
		}
	}

	for {
		// Check if we are doing a shutdown
		select {
		case <-ctx.Done():
			r.logger.Log(PrefixInfo, "Raft stopped")
			return
		default:
		}

		switch r.state.GetState() {
		case Follower:
			r.runAsFollower(ctx)
		case Candidate:
			r.runAsCandidate(ctx)
		case Leader:
			r.runAsLeader(ctx)
		}
	}
}

func (r *raft) runAsFollower(ctx context.Context) {
	r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))

	timer := time.NewTimer(electionTimeout)
	defer timer.Stop()

	for r.state.GetState() == Follower {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			r.state.SetState(Candidate)
			return

		case call := <-r.server.GetAppendEntriesCh():
			r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))
			if r.acceptAppendEntries(&call) {
				in := call.In.(schema.AppendEntriesInput)
				r.leaderId = in.LeaderId
				r.log.SaveNewEntries(in)
				r.state.SetTerm(in.Term)
				// reset counter only if leader is accepted
				timer.Reset(electionTimeout)
			}

		case call := <-r.server.GetRequestVoteCh():
			r.logger.Log(PrefixRequestVote)
			if r.acceptRequestVote(&call) {
				input := call.In.(schema.RequestVoteInput)
				r.state.SetVote(input.CandidateId)
			}

		case call := <-r.server.GetNewEntryCh():
			r.acceptNewEntry(&call)

		case call := <-r.server.GetLogCh():
			call.Respond(schema.Log{
				Term:    r.state.GetTerm(),
				State:   r.state.GetState().String(),
				Entries: r.log.GetEntriesFrom(0),
			}, nil)
		}
	}
}

func (r *raft) runAsCandidate(ctx context.Context) {
	r.state.IncrementTerm()
	r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))

	count := 0
	voteCh := r.sendRequestVote()

	for r.state.GetState() == Candidate {
		select {
		case <-ctx.Done():
			return

		case <-time.After(electionTimeout):
			return

		case resp := <-voteCh:
			if resp.Err != nil {
				r.logger.Log(PrefixError, resp.Err.Error())
				continue
			}

			vote, ok := resp.Val.(schema.RequestVoteOutput)
			if !ok {
				r.logger.Log(PrefixError, "wrong type received")
				continue
			}

			if vote.VoteGranted {
				r.logger.Log(PrefixInfo, "vote granted")
				count += 1
			}

			if majorityReached(count, r.clusterSize) {
				r.logger.Log(PrefixInfo, "majority reached")
				r.state.SetState(Leader)
				continue
			}

		case call := <-r.server.GetAppendEntriesCh():
			if r.acceptAppendEntries(&call) {
				in := call.In.(schema.AppendEntriesInput)
				r.leaderId = in.LeaderId
				r.log.SaveNewEntries(in)

				r.state.SetTerm(in.Term)
				r.state.SetState(Follower)
			}

		case call := <-r.server.GetRequestVoteCh():
			r.logger.Log(PrefixRequestVote)
			if r.acceptRequestVote(&call) {
				input := call.In.(schema.RequestVoteInput)
				r.state.SetVote(input.CandidateId)
			}

		case call := <-r.server.GetNewEntryCh():
			r.acceptNewEntry(&call)

		case call := <-r.server.GetLogCh():
			call.Respond(schema.Log{
				Term:    r.state.GetTerm(),
				State:   r.state.GetState().String(),
				Entries: r.log.GetEntriesFrom(0),
			}, nil)
		}
	}
}

func (r *raft) runAsLeader(ctx context.Context) {
	r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))

	nextIndexes := NewIndexesMap(r.clients, r.log)

	// this context is used for cancelling pushing new entries if the leader lost its status
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r.leaderId = r.config.ID

	timer := time.NewTimer(requestTimeout)
	defer timer.Stop()

	r.sendHeartbeat(ctx)

	for r.state.GetState() == Leader {
		select {
		case <-ctx.Done():
			return

		case <-timer.C: // heartbeat send
			r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))
			r.sendHeartbeat(ctx)
			timer.Reset(requestTimeout)

		case call := <-r.server.GetAppendEntriesCh():
			if r.acceptAppendEntries(&call) {
				in := call.In.(schema.AppendEntriesInput)
				r.leaderId = in.LeaderId
				r.log.SaveNewEntries(in)

				r.state.SetTerm(in.Term)
				r.state.SetState(Follower)
			}

		case call := <-r.server.GetRequestVoteCh():
			r.logger.Log(PrefixRequestVote)
			if r.acceptRequestVote(&call) {
				input := call.In.(schema.RequestVoteInput)
				r.state.SetVote(input.CandidateId)
			}

		case call := <-r.server.GetNewEntryCh():
			if r.acceptNewEntry(&call) {
				cmd := call.In.(schema.Command)
				r.logger.Log(PrefixInfo, "New entry:", cmd)
				r.log.AppendEntries([]*schema.Entry{{Term: r.state.GetTerm(), Data: &cmd}})
				r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))

				count := <-r.sendAppendEntries(ctx, nextIndexes)
				if majorityReached(count, r.clusterSize) {
					index, _ := r.log.LastEntryDecomposed()
					r.log.SetCommitIndex(index)
					r.logger.Log(PrefixInfo, "Commit index increased")
					r.logger.Log(PrefixInfo, fmt.Sprintf("%v", r.Stats()))
				}
			}

		case call := <-r.server.GetLogCh():
			call.Respond(schema.Log{
				Term:    r.state.GetTerm(),
				State:   r.state.GetState().String(),
				Entries: r.log.GetEntriesFrom(0),
			}, nil)
		}
	}
}
