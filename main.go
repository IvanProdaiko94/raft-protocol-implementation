package main

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	"github.com/IvanProdaiko94/raft-protocol-implementation/server"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type specification struct {
	ID int32 `envconfig:"ID" required:"true"`
}

func main() {
	var spec specification
	envconfig.MustProcess("", &spec)

	var nodeConfig = env.Cluster[spec.ID]
	var otherNodes = append(env.Cluster[0:spec.ID], env.Cluster[spec.ID:]...)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	var node = server.New(nodeConfig)
	go func() {
		log.Printf("\nStarting gRPC server at address: %s\n", nodeConfig.Address)
		err := node.Launch()
		if err != nil {
			panic(err)
		}
	}()

	err := node.InitClients(otherNodes)
	if err != nil {
		panic(err)
	}

	node.RunAsFollower(ctx)
	<-stop

	log.Println("Shutting down gRPC server")
	node.Stop()
}
