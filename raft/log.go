package raft

import (
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"sync"
)

type IndexedEntry struct {
	Index int32
	*schema.Entry
}

type Log struct {
	entries     []*schema.Entry
	lastApplied int32
	commitIndex int32
	sync.Mutex
}

func NewLog() *Log {
	return &Log{commitIndex: -1, lastApplied: -1}
}

func (l *Log) GetCommitIndex() int32 {
	return l.commitIndex
}

func (l *Log) SetCommitIndex(i int32) {
	l.commitIndex = i
}

func (l *Log) GetLastApplied() int32 {
	return l.lastApplied
}

func (l *Log) setLastApplied(la int32) {
	l.lastApplied = la
}

func (l *Log) Len() int32 {
	return int32(len(l.entries))
}

func (l *Log) IsEmpty() bool {
	return len(l.entries) == 0
}

func (l *Log) EntryByIndex(i int32) *IndexedEntry {
	if i > -1 && i < l.Len() {
		return &IndexedEntry{
			Index: i,
			Entry: l.entries[i],
		}
	}
	return nil
}

func (l *Log) EntryByIndexDecomposed(i int32) (int32, int32) {
	if i > -1 && i < l.Len() {
		entry := l.EntryByIndex(l.Len() - 1)
		return entry.Index, entry.Term
	}
	return -1, -1
}

func (l *Log) LastEntry() *IndexedEntry {
	if l.Len() > 0 {
		return l.EntryByIndex(l.Len() - 1)
	}
	return nil
}

func (l *Log) LastEntryDecomposed() (int32, int32) {
	if l.Len() > 0 {
		entry := l.EntryByIndex(l.Len() - 1)
		return entry.Index, entry.Term
	}
	return -1, -1
}

func (l *Log) AppendEntries(entries []*schema.Entry) {
	l.Lock()
	l.entries = append(l.entries, entries...)
	l.Unlock()
}

func (l *Log) IsCandidatesLogAtLeastAsUpToDateAsCurrent(lastLogTerm, lastLogIndex int32) bool {
	lastEntry := l.LastEntry()
	if lastEntry == nil { // empty log.
		return true
	}

	if lastEntry.Term != lastLogTerm {
		if lastLogTerm >= lastEntry.Term {
			return true
		}
	} else {
		if lastLogIndex >= lastEntry.Index {
			return true
		}
	}
	return false
}

func (l *Log) SaveNewEntries(in schema.AppendEntriesInput) {
	if len(in.Entries) == 0 {
		return
	}

	l.Lock()
	defer l.Unlock()

	i := in.PrevLogIndex + 1 // entry of local log
	// delete all entries that are in local log, but not in leader's one
	if in.PrevLogIndex < l.Len()-1 {
		l.entries = l.entries[:i]
	}

	defer func() {
		if in.LeaderCommit > l.GetCommitIndex() {
			commitIndex := in.LeaderCommit
			l.SetCommitIndex(commitIndex)
		}
	}()

	for j := 0; j < len(in.Entries); j++ { // j - entry of new log
		entry := l.EntryByIndex(i)

		if entry == nil { // no entry with this index (reached the end of the log)
			l.entries = append(l.entries, in.Entries[j:]...)
			return
		}

		if entry.Term != in.Entries[j].Term { // entry conflicts with one in leaders log
			l.entries = l.entries[:i-1]
			l.entries = append(l.entries, in.Entries[j:]...)
			return
		}

		i++
	}
}

func (l *Log) GetEntriesFrom(i int32) []*schema.Entry {
	l.Lock()
	defer l.Unlock()
	if i > -1 && i < l.Len() {
		return l.entries[i:]
	}
	return nil
}
