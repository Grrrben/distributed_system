package main

import (
	"context"
	"fmt"
	"github.com/grrrben/distsys/grades"
	"github.com/grrrben/distsys/log"
	"github.com/grrrben/distsys/registry"
	"github.com/grrrben/distsys/service"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	var r registry.Registration
	r.ServiceUrl = serviceAddress
	r.ServiceName = registry.GradingService
	r.RequiredServices = []registry.ServiceName{registry.LogService}
	r.ServiceUpdateUrl = r.ServiceUrl + "/services"
	r.HeartbeatUrl = r.ServiceUrl + "/heartbeat"

	ctx, err := service.Start(context.Background(), host, port, r, grades.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("Logging Service found at %v\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	<-ctx.Done()
	fmt.Println("Shutting down Grading Service")
}
