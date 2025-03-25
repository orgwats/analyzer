package kafka

import (
	"github.com/IBM/sarama"
)

type Consumer struct {
	sarama.Consumer
}

type OrderResult struct {
	OrderSide string // "BUY", "SELL"
	IsFilled  bool
}

func NewConsumer(addr string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{addr}, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer}, nil
}

func (c *Consumer) NewPartitionConsumer(topic string) (sarama.PartitionConsumer, error) {
	partitionConsumer, err := c.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err

	}
	return partitionConsumer, nil
}
