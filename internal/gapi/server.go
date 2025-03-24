package gapi

import (
	"github.com/google/uuid"
	"github.com/orgwats/analyzer/internal/analyzer"
	"github.com/orgwats/analyzer/internal/config"

	pb "github.com/orgwats/idl/gen/go/analyzer"
)

type Server struct {
	pb.UnimplementedAnalyzerServer

	// TODO: 임시
	cfg   *config.Config
	store interface{}

	analyzers map[uuid.UUID]*analyzer.Analyzer
}

func NewServer(cfg *config.Config, store interface{}) *Server {
	return &Server{
		cfg:       cfg,
		store:     store,
		analyzers: make(map[uuid.UUID]*analyzer.Analyzer),
	}
}
