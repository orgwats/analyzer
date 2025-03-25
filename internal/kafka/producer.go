package kafka

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	sarama.SyncProducer
}

type Order struct {
	OrderSide string // "BUY", "SELL"
	Key       string
	Symbol    string
	Size      string
}

func NewProducer(addr string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 모든 브로커 부터 응답 대기
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{addr}, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer}, nil
}

func (p *Producer) Send(analyzerId string, order *Order) (partition int32, offset int64, err error) {
	data, _ := json.Marshal(order)

	msg := &sarama.ProducerMessage{
		Topic: "order",
		Key:   sarama.StringEncoder(analyzerId),
		Value: sarama.StringEncoder(string(data)),
	}

	partition, offset, err = p.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	return
}
