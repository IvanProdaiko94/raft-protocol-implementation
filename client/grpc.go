package client

import (
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	schemapb "github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"google.golang.org/grpc"
	"log"
)

func CreateGRPC(address string) (schemapb.NodeClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, err
	}
	return schemapb.NewNodeClient(conn), nil
}

func CreateMultiple(nodes []env.Node) ([]schemapb.NodeClient, error) {
	var err error
	clients := make([]schemapb.NodeClient, len(nodes))
	for i, node := range nodes {
		clients[i], err = CreateGRPC(node.Address)
		if err != nil {
			return nil, err
		}
	}
	return clients, err
}
