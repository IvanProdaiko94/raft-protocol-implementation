package raft

import (
	"github.com/IvanProdaiko94/raft-protocol-implementation/rpc"
	"sync"
)

type Indexes struct {
	sync.Map
	sync.Mutex
}

func (ni *Indexes) SetIndex(clientId int32, i int32) {
	ni.Store(clientId, i)
}

func (ni *Indexes) DecrementIndex(clientId int32) bool {
	val, ok := ni.Load(clientId)
	if !ok {
		return false
	}
	prev, ok := val.(int32)
	if !ok || prev == 0 {
		return false
	}
	ni.Store(clientId, prev-1)
	return true
}

func (ni *Indexes) Reset(clients []rpc.Client, log *Log) {
	ni.Lock()
	defer ni.Unlock()
	for _, client := range clients {
		var i int32 = -1
		if entry := log.LastEntry(); entry != nil {
			i = entry.Index + 1
		}
		ni.Store(client.GetId(), i)
	}
}

func NewIndexesMap(clients []rpc.Client, log *Log) *Indexes {
	mp := &Indexes{}
	mp.Reset(clients, log)
	return mp
}
