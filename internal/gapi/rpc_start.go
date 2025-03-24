package gapi

import (
	"context"

	"github.com/orgwats/analyzer/internal/analyzer"
	aPb "github.com/orgwats/idl/gen/go/analyzer"
	mPb "github.com/orgwats/idl/gen/go/market"
)

func (s *Server) Start(ctx context.Context, req *aPb.StartAnalyzerRequest) (*aPb.StartAnalyzerResponse, error) {
	// addr := s.cfg.~~~
	addr := "localhost:50051"
	client, close, err := NewClient(mPb.NewMarketClient, addr)
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

	uuid, analyzer := analyzer.NewAnalyzer(ctx, s.store, client, stream, req.Symbol, close)
	s.analyzers[uuid] = analyzer

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
