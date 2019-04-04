package raft

import (
	"log"
	"os"
	"sync"
)

const (
	PrefixHeartbeat      = "[HEARTBEAT] "
	PrefixStateChanged   = "[STATE CHANGED] "
	PrefixTermChanged    = "[TERM CHANGED] "
	PrefixElection       = "[ELECTION] "
	PrefixElectionResult = "[ELECTION RESULT] "
	PrefixElectionStop   = "[ELECTION STOPPED] "
	PrefixInfo           = "[INFO] "
	PrefixError          = "[ERROR] "
	PrefixAppendEntries  = "[APPEND ENTRIES] "
	PrefixRequestVote    = "[REQUEST VOTE] "
)

type Logger struct {
	l *log.Logger
	sync.Mutex
}

func (logger *Logger) Log(prefix string, v ...interface{}) {
	logger.Lock()
	defer logger.Unlock()
	if prefix == PrefixError {
		logger.l.SetOutput(os.Stderr)
	}
	logger.l.SetPrefix(prefix)
	logger.l.Println(v...)
	logger.l.SetOutput(os.Stdout)
}

func NewLogger() *Logger {
	return &Logger{
		l: log.New(os.Stdout, "[INIT]", log.Ltime),
	}
}
