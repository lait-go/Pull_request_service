package kafkaconsumer

import (
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string `envconfig:"KAFKA_BROKERS" required:"true"`
	Topic   string   `envconfig:"KAFKA_TOPIC" required:"true"`
}

type Consumer struct {
	Reader   *kafka.Reader
}

func NewConsumer(cfg Config) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  cfg.Brokers,
			Topic:    cfg.Topic,
			GroupID:  "my-group",
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}
