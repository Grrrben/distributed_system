package main

import (
	"context"
	"fmt"
	"github.com/grrrben/distsys/registry"
	"log"
	"net/http"
)

func main() {
	registry.SetUpHeartBeatCheck()

	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println("Registry Service started. Press any key to quit")

		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Shutting down Registry Service")
}
