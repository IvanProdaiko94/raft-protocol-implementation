package rpc

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	"github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"log"
)

type Client interface {
	Deal() error
	Close()
	GetId() int32
	schema.NodeClient
}

type client struct {
	address   string
	id        int32
	transport schema.NodeClient
	conn      *grpc.ClientConn
}

func (c *client) Deal() error {
	conn, err := grpc.Dial(c.address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return err
	}
	c.transport = schema.NewNodeClient(conn)
	c.conn = conn
	return nil
}

func (c *client) Close() {
	_ = c.conn.Close()
}

func (c *client) RequestVote(ctx context.Context, in *schema.RequestVoteInput, opts ...grpc.CallOption) (*schema.RequestVoteOutput, error) {
	return c.transport.RequestVote(ctx, in, opts...)
}

func (c *client) AppendEntries(ctx context.Context, in *schema.AppendEntriesInput, opts ...grpc.CallOption) (*schema.AppendEntriesOutput, error) {
	return c.transport.AppendEntries(ctx, in, opts...)
}

func (c *client) NewEntry(ctx context.Context, in *schema.Command, opts ...grpc.CallOption) (*schema.PushResponse, error) {
	return c.transport.NewEntry(ctx, in, opts...)
}

func (c *client) GetLog(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*schema.Log, error) {
	return c.transport.GetLog(ctx, in, opts...)
}

func (c *client) GetId() int32 {
	return c.id
}

func NewClient(id int32, address string) Client {
	return &client{
		id:      id,
		address: address,
	}
}

func NewClients(nodes []env.Node) []Client {
	clients := make([]Client, len(nodes))
	for i, node := range nodes {
		clients[i] = NewClient(node.ID, node.Address)
	}
	return clients
}
