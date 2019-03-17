package log

import (
	"encoding/json"
	"sync"
)

type Entry struct {
	Term int
	data map[string]interface{}
}

func EntryFromBytes(data []byte, term int) (*Entry, error) {
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

func (l *Log) LastIndex() int {
	l.Lock()
	defer l.Unlock()
	return len(l.entries) - 1
}

func (l *Log) EntryByIndex(i int) *IndexedEntry {
	lastIndex := l.LastIndex()
	if lastIndex > i {
		return nil
	}
	return &IndexedEntry{
		Index: i,
		Entry: l.entries[i],
	}
}

func (l *Log) LastEntry() *IndexedEntry {
	lastIndex := l.LastIndex()
	return l.EntryByIndex(lastIndex)
}

func (l *Log) AppendEntries(entries []Entry) {
	l.Lock()
	l.entries = append(l.entries, entries...)
	l.Unlock()
}

func New() *Log {
	return &Log{}
}
