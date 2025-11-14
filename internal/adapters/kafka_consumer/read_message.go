package kafkaconsumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
    msg, err := c.Reader.ReadMessage(ctx)
    if err != nil {
        return kafka.Message{}, err
    }
    return msg, nil
}