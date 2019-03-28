package client

import (
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

func CreateMultiple(addressList []string) ([]schemapb.NodeClient, error) {
	var err error
	clients := make([]schemapb.NodeClient, len(addressList))
	for i, address := range addressList {
		clients[i], err = CreateGRPC(address)
		if err != nil {
			return nil, err
		}
	}
	return clients, err
}
