package main

import (
	"context"
	"fmt"
	"github.com/grrrben/distsys/log"
	"github.com/grrrben/distsys/registry"
	"github.com/grrrben/distsys/service"
	stlog "log"
)

func main() {
	log.Run("./app.log")

	host, port := "localhost", "4000"

	var reg registry.Registration
	reg.ServiceName = registry.LogService
	reg.ServiceUrl = fmt.Sprintf("http://%s:%s", host, port)
	reg.RequiredServices = make([]registry.ServiceName, 0)
	reg.ServiceUpdateUrl = reg.ServiceUrl + "/services"
	reg.HeartbeatUrl = reg.ServiceUrl + "/heartbeat"

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		reg,
		log.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatalf("Fatal when starting logservice: %s", err)
	}

	<-ctx.Done() // waiting for the cancel of the ctx in the service
	fmt.Println("Shutting down log service")
}
