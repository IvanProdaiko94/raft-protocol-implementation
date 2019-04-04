package raft

import (
	"fmt"
)

func (r *raft) Stats() map[string]string {
	toString := func(v int32) string {
		return fmt.Sprintf("%d", v)
	}
	lastLogIndex, lastLogTerm := r.log.LastEntryDecomposed()
	return map[string]string{
		"state":          r.state.GetState().String(),
		"term":           toString(r.state.GetTerm()),
		"last_log_index": toString(lastLogIndex),
		"last_log_term":  toString(lastLogTerm),
		"commit_index":   toString(r.log.GetCommitIndex()),
		"vote_id":        toString(r.state.GetVote()),
		"leader_id":      toString(r.leaderId),
	}
}
