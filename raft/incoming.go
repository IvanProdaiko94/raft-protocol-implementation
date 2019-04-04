package raft

import (
	"fmt"
	"github.com/IvanProdaiko94/raft-protocol-implementation/rpc"
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"github.com/pkg/errors"
)

func (r *raft) acceptAppendEntries(call *rpc.Call) bool {
	in, ok := call.In.(schema.AppendEntriesInput)
	if !ok {
		call.Respond(nil, errors.New("Failed to cast input"))
		return false
	}
	success := r.validateAppendEntries(in)
	call.Respond(schema.AppendEntriesOutput{Term: r.state.GetTerm(), Success: success}, nil)
	return success
}

func (r *raft) acceptRequestVote(call *rpc.Call) bool {
	in, ok := call.In.(schema.RequestVoteInput)
	if !ok {
		call.Respond(nil, errors.New("Failed to cast input"))
		return false
	}
	voteGranted := r.validateRequestVote(in)
	call.Respond(schema.RequestVoteOutput{Term: r.state.GetTerm(), VoteGranted: voteGranted}, nil)
	return voteGranted
}

func (r *raft) acceptNewEntry(call *rpc.Call) bool {
	_, ok := call.In.(schema.Command)
	if !ok {
		call.Respond(nil, errors.New("Failed to cast input"))
		return false
	}
	accept := r.validateIsLeader()
	call.Respond(schema.PushResponse{Success: accept, Leader: r.leaderId}, nil)
	return accept
}

func (r *raft) validateAppendEntries(in schema.AppendEntriesInput) bool {
	if in.Term < r.state.GetTerm() {
		return false
	}

	if len(in.Entries) > 0 {
		r.logger.Log(PrefixInfo, "received new entries: ", in.Entries)
		entry := r.log.EntryByIndex(in.PrevLogIndex)
		if entry != nil && entry.Term != in.PrevLogTerm {
			return false
		}
	}

	return true
}

func (r *raft) validateRequestVote(in schema.RequestVoteInput) bool {
	if in.Term < r.state.GetTerm() {
		return false
	}

	if r.state.VotedInThisTerm() {
		return false
	}

	if r.state.GetVote() == in.CandidateId {
		return false
	}

	upToDate := r.log.IsCandidatesLogAtLeastAsUpToDateAsCurrent(in.LastLogTerm, in.LastLogIndex)
	fmt.Println("Candidates log is up to date", upToDate)
	return upToDate
}

func (r *raft) validateIsLeader() bool {
	return r.leaderId == r.config.ID && r.state.GetState() == Leader
}

func majorityReached(count, nodesNumber int) bool {
	return count >= (nodesNumber/2)+1
}
