package raft

import (
	"context"
	"fmt"
	"github.com/IvanProdaiko94/raft-protocol-implementation/rpc"
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
)

func (r *raft) sendRequestVote() chan rpc.Resp {
	ch := make(chan rpc.Resp)
	for _, client := range r.clients {
		go func(client rpc.Client) {
			lastLogIndex, lastLogTerm := r.log.LastEntryDecomposed()
			ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
			output, err := client.RequestVote(ctx, &schema.RequestVoteInput{
				Term:         r.state.GetTerm(),
				CandidateId:  r.config.ID,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			})
			if err != nil {
				ch <- rpc.Resp{Val: nil, Err: err}
				return
			}
			ch <- rpc.Resp{Val: *output, Err: nil}
		}(client)
	}

	go func() {
		ch <- rpc.Resp{Val: schema.RequestVoteOutput{
			Term:        r.state.GetTerm(),
			VoteGranted: true,
		}, Err: nil}
	}()

	return ch
}

func (r *raft) sendHeartbeat(ctx context.Context) {
	for _, client := range r.clients {
		go func(client rpc.Client) {
			prevLogIndex, prevLogTerm := r.log.EntryByIndexDecomposed(r.log.Len() - 1)
			out, err := client.AppendEntries(ctx, &schema.AppendEntriesInput{
				Term:         r.state.GetTerm(),
				LeaderId:     r.config.ID,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      nil,
			})
			if err != nil { // continue till entry actually be saved to the node.
				r.logger.Log(PrefixError, err.Error())
				return
			}

			if out.Term > r.state.GetTerm() {
				r.state.SetTerm(out.Term)
				r.state.SetState(Follower)
				return
			}

			if !out.Success {
				r.logger.Log(
					PrefixError,
					fmt.Sprintf("heartbeat is not successful. Remote term: %d", out.Term),
				)
			}
		}(client)
	}
}

func (r *raft) sendAppendEntries(ctx context.Context, ni *Indexes) chan int {
	ch := make(chan int)
	count := 1
	for _, client := range r.clients {
		go func(client rpc.Client) {
			lastLogIndex, _ := r.log.LastEntryDecomposed()

			tmp, ok := ni.Load(client.GetId())
			if !ok {
				r.logger.Log(PrefixError, fmt.Sprintf("next index not found for client %d", client.GetId()))
				return
			}
			nextIndex, ok := tmp.(int32)
			if !ok {
				r.logger.Log(PrefixError, fmt.Sprintf("next index not correct type for client %d", client.GetId()))
				return
			}

			if nextIndex == -1 {
				nextIndex = 0
			}

			if lastLogIndex < nextIndex { // all entries are sent already
				return
			}

			var entries = r.log.GetEntriesFrom(nextIndex)

			success := false
			for !success {
				select {
				case <-ctx.Done():
					return
				default:
					var (
						prevLogIndex int32 = 0 // must be updated according to next index
						prevLogTerm  int32 = 0 // must be updated according to next index
					)
					prevLogEntry := r.log.EntryByIndex(lastLogIndex - 1)
					if prevLogEntry != nil {
						prevLogIndex = prevLogEntry.Index
						prevLogTerm = prevLogEntry.Term
					}
					out, err := client.AppendEntries(ctx, &schema.AppendEntriesInput{
						Term:         r.state.GetTerm(),
						LeaderId:     r.config.ID,
						LeaderCommit: r.log.GetCommitIndex(),
						PrevLogIndex: prevLogIndex,
						PrevLogTerm:  prevLogTerm,
						Entries:      entries,
					})
					if err != nil { // continue till entry actually be saved to the node.
						r.logger.Log(PrefixError, err.Error())
						continue
					}
					if out.Success {
						success = true
						ni.Store(client.GetId(), nextIndex+int32(len(entries)))
						count += 1
						ch <- count
					}
					if out.Term > r.state.GetTerm() {
						r.state.SetState(Follower)
						return
					}

					if !out.Success {
						r.logger.Log(PrefixError, fmt.Sprintf("append entries failed. Remote term: %d", out.Term))
						nextIndex -= 1 // If AppendEntries fails because of log inconsistency: decrement nextIndex and retry
					}
				}
			}
		}(client)
	}
	return ch
}
