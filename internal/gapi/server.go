package gapi

import (
	"encoding/json"
	"log"

	"github.com/orgwats/analyzer/internal/analyzer"
	"github.com/orgwats/analyzer/internal/config"
	"github.com/orgwats/analyzer/internal/kafka"

	pb "github.com/orgwats/idl/gen/go/analyzer"
)

type Server struct {
	pb.UnimplementedAnalyzerServer

	// TODO: 임시
	cfg   *config.Config
	store interface{}

	analyzers map[string]*analyzer.Analyzer
}

func NewServer(cfg *config.Config, store interface{}) *Server {
	return &Server{
		cfg:       cfg,
		store:     store,
		analyzers: make(map[string]*analyzer.Analyzer),
	}
}

func (s *Server) Run() {
	// s.cfg.~~~
	addr := "kafka:9093"
	consumer, err := kafka.NewConsumer(addr)
	if err != nil {
		log.Fatal("cannot create kafka consumer:", err)
	}

	// s.cfg.~~~
	topic := "order-result"
	partitionConsumer, err := consumer.NewPartitionConsumer(topic)
	if err != nil {
		log.Fatal("cannot create partitionConsumer:", err)
	}

	go func() {
		for {
			select {
			case err := <-partitionConsumer.Errors():
				log.Printf("Consumer error: %v", err)
			case msg := <-partitionConsumer.Messages():
				var orderResult kafka.OrderResult
				err := json.Unmarshal(msg.Value, &orderResult)
				if err != nil {
					log.Println("cannot unmarshal kafka message")
				}

				log.Printf("analyzerId : %s / result : %s|%t", msg.Key, orderResult.OrderSide, orderResult.IsFilled)

				if orderResult.IsFilled {
					s.analyzers[string(msg.Key)].Position.IsFilled = true
				} else {
					s.analyzers[string(msg.Key)].Position.Type = ""
					s.analyzers[string(msg.Key)].Position.Price = 0
					s.analyzers[string(msg.Key)].Position.TakeProfit = 0
					s.analyzers[string(msg.Key)].Position.StopLoss = 0
					s.analyzers[string(msg.Key)].Position.IsFilled = false
				}
			}
		}
	}()
}
