package main

import (
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/orgwats/analyzer/internal/config"
	"github.com/orgwats/analyzer/internal/gapi"
	aPb "github.com/orgwats/idl/gen/go/analyzer"
)

func main() {
	// TODO 임시
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// TODO 임시
	store := struct{}{}
	server := gapi.NewServer(cfg, store)
	grpcServer := grpc.NewServer()

	aPb.RegisterAnalyzerServer(grpcServer, server)

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("cannot listen network address:", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("start analyzer GRPC server at %s", listener.Addr().String())

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("market server failed to serve:", err)
		}
	}()
	wg.Wait()
}
