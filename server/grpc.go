package server

import (
	schemapb "github.com/IvanProdaiko94/raft-protocol-implementation/schema"
	"google.golang.org/grpc"
	"log"
	"net"
)

func CreateGRPC(n *Node) *grpc.Server {
	server := grpc.NewServer()
	schemapb.RegisterNodeServer(server, n)
	return server
}

func Listen(server *grpc.Server, addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return server.Serve(listen)
}
