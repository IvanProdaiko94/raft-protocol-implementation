package raft

import "sync/atomic"

const (
	Follower = S(iota)
	Candidate
	Leader
)

type S int32

func (s S) String() string {
	if s == Follower {
		return "Follower"
	} else if s == Candidate {
		return "Candidate"
	} else if s == Leader {
		return "Leader"
	}
	panic("incorrect state")
}

type State struct {
	state    int32
	t        int32
	votedFor int32
}

func (s *State) SetState(newState S) {
	atomic.StoreInt32(&s.state, int32(newState))
}

func (s *State) GetState() S {
	return S(atomic.LoadInt32(&s.state))
}

func (s *State) SetVote(candidateId int32) {
	atomic.StoreInt32(&s.votedFor, candidateId)
}

func (s *State) GetVote() int32 {
	return atomic.LoadInt32(&s.votedFor)
}

func (s *State) VotedInThisTerm() bool {
	return atomic.LoadInt32(&s.votedFor) > -1
}

func (s *State) IncrementTerm() {
	atomic.AddInt32(&s.t, 1)
	atomic.StoreInt32(&s.votedFor, -1)
}

func (s *State) SetTerm(new int32) {
	atomic.StoreInt32(&s.t, new)
	if s.GetTerm() != new {
		atomic.StoreInt32(&s.votedFor, -1)
	}
}

func (s *State) GetTerm() int32 {
	return atomic.LoadInt32(&s.t)
}

func NewState() *State {
	return &State{t: 0, votedFor: -1}
}
