package main

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/env"
	"github.com/IvanProdaiko94/raft-protocol-implementation/raft"
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
	var otherNodes = append(env.Cluster[0:spec.ID], env.Cluster[spec.ID+1:]...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	protocol := raft.New(nodeConfig, otherNodes, len(env.Cluster))

	go protocol.Start(ctx)
	defer protocol.Stop()

	<-stop

	log.Println("Shutting down gRPC server")
}
