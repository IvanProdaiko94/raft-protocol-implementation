package log

import (
	"encoding/json"
	"github.com/golang/protobuf/ptypes/any"
	"sync"
)

type Entry struct {
	Term int32
	data map[string]interface{}
}

func EntryFromBytes(data []byte, term int32) (*Entry, error) {
	var newEntry map[string]interface{}
	err := json.Unmarshal(data, &newEntry)
	if err != nil {
		return nil, err
	}
	return &Entry{
		Term: term,
		data: newEntry,
	}, nil
}

type IndexedEntry struct {
	Index int
	Entry
}

type Log struct {
	entries []Entry
	sync.Mutex
}

func (l *Log) lastIndex() int {
	return len(l.entries) - 1
}

func (l *Log) EntryByIndex(i int) *IndexedEntry {
	lastIndex := l.lastIndex()
	if lastIndex > i || lastIndex == -1 {
		return nil
	}
	return &IndexedEntry{
		Index: i,
		Entry: l.entries[i],
	}
}

func (l *Log) LastEntry() *IndexedEntry {
	lastIndex := l.lastIndex()
	if lastIndex > -1 {
		return l.EntryByIndex(lastIndex)
	}
	return nil
}

func (l *Log) AppendEntries(entries []Entry) {
	l.Lock()
	l.entries = append(l.entries, entries...)
	l.Unlock()
}

func (l *Log) IsEmpty() bool {
	return len(l.entries) == 0
}

func (l *Log) SolveConflicts(prevLogTerm, prevLogIndex int32, newEntries []*any.Any) {

}

func New() *Log {
	return &Log{}
}
