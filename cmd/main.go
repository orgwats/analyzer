package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
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

		server.Run()

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("market server failed to serve:", err)
		}
	}()

	port := ":8080"

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		var req aPb.StartAnalyzerRequest
		// 요청 본문에서 JSON을 파싱하여 req에 저장
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// req.Symbol에 body에서 받은 값이 들어있음
		rsp, err := server.Start(context.Background(), &req)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to start analyzer", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(rsp)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		rsp, _ := server.Stop(context.Background(), &aPb.StopAnalyzerRequest{Symbol: "ETHUSDT"})
		data, _ := json.Marshal(rsp)
		w.Write([]byte(data))
	})
	log.Printf("start analyzer REST server at %s", port)
	log.Fatal(http.ListenAndServe(port, nil))

	wg.Wait()
}
