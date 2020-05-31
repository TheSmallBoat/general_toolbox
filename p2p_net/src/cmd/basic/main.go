package main

import (
	"os"
	"os/signal"

	"awesomeProject/beacon/p2p_network/core_module"
	"awesomeProject/beacon/p2p_network/libs/kademlia"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.PanicLevel))
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	node, err := core_module.NewNode(core_module.WithNodeLogger(logger), core_module.WithNodeBindPort(9000))
	if err != nil {
		panic(err)
	}
	defer node.Close()

	overlay := kademlia.New()
	node.Bind(overlay.Protocol())

	if err := node.Listen(); err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
