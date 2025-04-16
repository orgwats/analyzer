package gapi

import (
	"context"

	"github.com/orgwats/analyzer/internal/analyzer"
	"github.com/orgwats/analyzer/internal/kafka"
	aPb "github.com/orgwats/idl/gen/go/analyzer"
	mPb "github.com/orgwats/idl/gen/go/market"
)

func (s *Server) Start(ctx context.Context, req *aPb.StartAnalyzerRequest) (*aPb.StartAnalyzerResponse, error) {
	// addr := s.cfg.~~~
	rpcAddr := "market-service:50051"
	kafkaAddr := "34.132.215.86:9092"

	client, close, err := NewClient(mPb.NewMarketClient, rpcAddr)
	if err != nil {
		close()
		return &aPb.StartAnalyzerResponse{
			Success: false,
			Message: "",
		}, err
	}

	stream, err := client.NewCandleStream(ctx, &mPb.NewCandleStreamRequest{Symbol: req.Symbol})
	if err != nil {
		close()
		return &aPb.StartAnalyzerResponse{
			Success: false,
			Message: "",
		}, err
	}

	producer, err := kafka.NewProducer(kafkaAddr)
	if err != nil {
		close()
		return &aPb.StartAnalyzerResponse{
			Success: false,
			Message: "",
		}, err
	}

	id, analyzer := analyzer.NewAnalyzer(ctx, s.store, client, stream, producer, req.Symbol, close)
	s.analyzers[id] = analyzer

	err = analyzer.Init()
	if err != nil {
		close()
		return &aPb.StartAnalyzerResponse{
			Success: false,
			Message: "",
		}, err
	}

	go analyzer.Run()

	return &aPb.StartAnalyzerResponse{
		Success: true,
		Message: "",
	}, nil
}
