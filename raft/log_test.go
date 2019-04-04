package raft

import (
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLog_GetEntriesFrom(t *testing.T) {
	log := Log{
		entries: nil,
	}
	res := log.GetEntriesFrom(10)
	assert.Equal(t, len(res), 0)
}

func TestLog_GetEntriesFrom2(t *testing.T) {
	log := Log{entries: []*schema.Entry{{}, {}, {}, {}}}

	res := log.GetEntriesFrom(2)
	assert.Equal(t, len(res), 2)
}

func TestLog_IsCandidatesLogAtLeastAsUpToDateAsCurrent(t *testing.T) {
	log := Log{entries: []*schema.Entry{{Term: 0}, {Term: 1}}}

	res := log.IsCandidatesLogAtLeastAsUpToDateAsCurrent(1, 1)
	assert.Equal(t, res, true)
}

func TestLog_IsCandidatesLogAtLeastAsUpToDateAsCurrent2(t *testing.T) {
	log := Log{entries: []*schema.Entry{{Term: 0}, {Term: 2}}}

	res := log.IsCandidatesLogAtLeastAsUpToDateAsCurrent(1, 1)
	assert.Equal(t, res, false)
}

func TestLog_IsCandidatesLogAtLeastAsUpToDateAsCurrent3(t *testing.T) {
	log := Log{entries: []*schema.Entry{{Term: 5}}}

	res := log.IsCandidatesLogAtLeastAsUpToDateAsCurrent(1, 1)
	assert.Equal(t, res, false)
}

func TestLog_SaveNewEntries(t *testing.T) { //case A Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.LastEntry().Index,
		PrevLogTerm:  leaderLog.LastEntry().Term,
		Entries:      leaderLog.GetEntriesFrom(leaderLog.Len() - 1),
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}

func TestLog_SaveNewEntries2(t *testing.T) { //case B Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.LastEntry().Index,
		PrevLogTerm:  leaderLog.LastEntry().Term,
		Entries:      leaderLog.GetEntriesFrom(4),
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}

func TestLog_SaveNewEntries3(t *testing.T) { //case C Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.LastEntry().Index,
		PrevLogTerm:  leaderLog.LastEntry().Term,
		Entries:      nil,
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}

func TestLog_SaveNewEntries4(t *testing.T) { //case D Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
			{Term: 7}, {Term: 7},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.LastEntry().Index,
		PrevLogTerm:  leaderLog.LastEntry().Term,
		Entries:      nil,
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}

func TestLog_SaveNewEntries5(t *testing.T) { //case E Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4}, {Term: 4}, {Term: 4},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.EntryByIndex(4).Index,
		PrevLogTerm:  leaderLog.EntryByIndex(4).Term,
		Entries:      leaderLog.GetEntriesFrom(5),
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}

func TestLog_SaveNewEntries6(t *testing.T) { //case F Figure 7
	leaderLog := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 4}, {Term: 4},
			{Term: 5}, {Term: 5},
			{Term: 6}, {Term: 6}, {Term: 6},
		},
	}

	l := Log{
		entries: []*schema.Entry{
			{Term: 1}, {Term: 1}, {Term: 1},
			{Term: 2}, {Term: 2}, {Term: 2},
			{Term: 3}, {Term: 3}, {Term: 3}, {Term: 3}, {Term: 3},
		},
	}

	l.SaveNewEntries(schema.AppendEntriesInput{
		Term:         8,
		PrevLogIndex: leaderLog.EntryByIndex(2).Index,
		PrevLogTerm:  leaderLog.EntryByIndex(2).Term,
		Entries:      leaderLog.GetEntriesFrom(3),
	})

	assert.Equal(t, leaderLog.entries, l.entries)
}
