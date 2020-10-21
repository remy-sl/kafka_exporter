package kafka_exporter

import (
	"context"

	"github.com/adambabik/kafka_exporter/pkg/types"
)

type KafkaClient interface {
	Brokers(ctx context.Context) ([]types.Broker, error)
	Topics(ctx context.Context) ([]types.Topic, error)
	Offsets(ctx context.Context) ([]types.Offset, error)
	ConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error)
	OffsetsForConsumerGroups(ctx context.Context) ([]types.Offset, error)
}
