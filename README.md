# raft-protocol-implementation
Raft protocol implementation as part of "Distributed databases" course in UCU

## Raft decomposes the consensus problem into three relatively independent subproblems:
- **Leader election**: a new leader must be chosen when an existing leader fails.
- **Log replication**: the leader must accept log entries
- **Safety**: the key safety property for Raft is the State Machine Safety Property. If any server has applied a particular 
log entry to its state machine, then no other server may apply a different command for the same log index. The solution involves an
additional restriction on the election mechanism.

##  Raft guarantees that each of these properties is true at all times.
- **Election Safety**: at most one leader can be elected in a given term.
- **Leader Append-Only**: a leader never overwrites or deletes entries in its log; it only appends new entries.
- **Log Matching**: if two logs contain an entry with the same index and term, then the logs are identical in all entries
up through the given index.
- **Leader Completeness**: if a log entry is committed in a given term, then that entry will be present in the logs
of the leaders for all higher-numbered terms.
- **State Machine Safety**: if a server has applied a log entry at a given index to its state machine, no other server
will ever apply a different log entry for the same index.

## Info about implementation
- Each server launches rpc server
- Communication with client is done via [raft http client](https://github.com/IvanProdaiko94/raft-protocol-implementation-client) or you can do it directly via [evans](https://github.com/ktr0731/evans)
- Creating new entry is done via leader node, other nodes will respond with no success and leader id.
- Each node could reveil log via `GetLog` rpc call. (this is done according to task requirements)

## How to launch:

- Install dependencies with `dep ensure`
- Change number of nodes and set an unique ID to each in `env/env.go`
- Run `main.go` with `ID=x` env variable set.
- Run several processes to create cluster.