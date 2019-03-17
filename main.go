package main

import (
	"context"
	"github.com/IvanProdaiko94/raft-protocol-implementation/server"
	"github.com/dvln/out"
	"github.com/kelseyhightower/envconfig"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type specification struct {
	NodesList   string `envconfig:"NODES_LIST" required:"true"`
	NodeAddress string `envconfig:"NODE_ADDRESS" required:"true"`
}

func main() {
	var spec specification
	envconfig.MustProcess("", &spec)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	var node = server.New()
	go func() {
		out.Infof("Starting gRPC server at address: %s\n", spec.NodeAddress)
		err := node.Launch(spec.NodeAddress)
		if err != nil {
			panic(err)
		}
		err = node.InitClients(strings.Split(spec.NodesList, ","))
		if err != nil {
			panic(err)
		}
	}()

	node.RunAsFollower(ctx)
	<-stop

	out.Infoln("Shutting down gRPC server")
	node.Stop()
}
