package rpc

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"net"
)

type Server interface {
	GetAppendEntriesCh() chan Call
	GetRequestVoteCh() chan Call
	GetNewEntryCh() chan Call
	GetLogCh() chan Call
	Start(context.Context) error
	Stop()
	schema.NodeServer
}

type server struct {
	config          env.Node
	appendEntriesCh chan Call
	requestVoteCh   chan Call
	newEntryCh      chan Call
	logCh           chan Call
	server          *grpc.Server
}

func (s *server) Stop() {
	if s.server != nil {
		s.server.Stop()
	}
}

func (s *server) GetAppendEntriesCh() chan Call {
	return s.appendEntriesCh
}

func (s *server) GetRequestVoteCh() chan Call {
	return s.requestVoteCh
}

func (s *server) GetNewEntryCh() chan Call {
	return s.newEntryCh
}

func (s *server) GetLogCh() chan Call {
	return s.logCh
}

func (s *server) AppendEntries(ctx context.Context, input *schema.AppendEntriesInput) (*schema.AppendEntriesOutput, error) {
	ch := make(chan Resp)
	defer close(ch)

	s.appendEntriesCh <- Call{In: *input, C: ch}
	// wait for request validation
	output := <-ch
	if output.Val == nil {
		return nil, output.Err
	}
	out := output.Val.(schema.AppendEntriesOutput)
	return &out, output.Err
}

func (s *server) RequestVote(ctx context.Context, input *schema.RequestVoteInput) (*schema.RequestVoteOutput, error) {
	ch := make(chan Resp)
	defer close(ch)

	s.requestVoteCh <- Call{In: *input, C: ch}
	output := <-ch
	if output.Val == nil {
		return nil, output.Err
	}
	out := output.Val.(schema.RequestVoteOutput)
	return &out, output.Err
}

func (s *server) NewEntry(ctx context.Context, input *schema.Command) (*schema.PushResponse, error) {
	ch := make(chan Resp)
	defer close(ch)

	s.newEntryCh <- Call{In: *input, C: ch}
	output := <-ch
	if output.Val == nil {
		return nil, output.Err
	}
	out := output.Val.(schema.PushResponse)
	return &out, output.Err
}

func (s *server) GetLog(ctx context.Context, input *empty.Empty) (*schema.Log, error) {
	ch := make(chan Resp)
	defer close(ch)

	s.logCh <- Call{In: *input, C: ch}
	output := <-ch
	if output.Val == nil {
		return nil, output.Err
	}
	out := output.Val.(schema.Log)
	return &out, output.Err
}

func (s *server) Start(ctx context.Context) error {
	server := grpc.NewServer()
	schema.RegisterNodeServer(server, s)
	listen, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	s.server = server
	return server.Serve(listen)
}

func NewServer(config env.Node) Server {
	return &server{
		config:          config,
		appendEntriesCh: make(chan Call),
		requestVoteCh:   make(chan Call),
		newEntryCh:      make(chan Call),
		logCh:           make(chan Call),
	}
}
