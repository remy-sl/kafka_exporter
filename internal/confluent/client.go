// +build cgo

package confluent

import (
	"context"
	"time"

	"github.com/adambabik/kafka_exporter/pkg/types"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type ConfigMap = kafka.ConfigMap

type Client struct {
	cfg ConfigMap
}

func New(cfg ConfigMap) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Brokers(ctx context.Context) (result []types.Broker, err error) {
	admin, err := kafka.NewAdminClient(&c.cfg)
	if err != nil {
		return nil, err
	}
	defer admin.Close()

	meta, err := admin.GetMetadata(nil, false, timeoutMsFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	result = make([]types.Broker, 0, len(meta.Brokers))
	for _, b := range meta.Brokers {
		result = append(result, types.Broker{
			ID:   int(b.ID),
			Host: b.Host,
			Port: b.Port,
		})
	}
	return result, nil
}

func (c *Client) Topics(ctx context.Context) (result []types.Topic, err error) {
	admin, err := kafka.NewAdminClient(&c.cfg)
	if err != nil {
		return nil, err
	}
	defer admin.Close()

	meta, err := admin.GetMetadata(nil, true, timeoutMsFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	result = make([]types.Topic, 0, len(meta.Topics))
	for _, t := range meta.Topics {
		partitions := make([]types.Partition, 0, len(t.Partitions))
		for _, p := range t.Partitions {
			partitions = append(partitions, types.Partition{
				ID:       int(p.ID),
				Leader:   int(p.Leader),
				Replicas: fromInt32toInt(p.Replicas),
				Isrs:     fromInt32toInt(p.Isrs),
			})
		}
		result = append(result, types.Topic{
			Name:       t.Topic,
			Partitions: partitions,
		})
	}
	return result, nil
}

func (c *Client) Offsets(ctx context.Context) (result []types.Offset, err error) {
	return
}

func (c *Client) ConsumerGroups(ctx context.Context) (result []types.ConsumerGroup, err error) {
	return
}

func (c *Client) OffsetsForConsumerGroups(ctx context.Context) (result []types.Offset, err error) {
	return
}

func fromInt32toInt(s []int32) []int {
	result := make([]int, 0, len(s))
	for _, item := range s {
		result = append(result, int(item))
	}
	return result
}

func timeoutMsFromCtx(ctx context.Context) int {
	timeout := time.Second * 10
	if deadline, ok := ctx.Deadline(); ok {
		timeout = deadline.Sub(time.Now())
	}
	return int(timeout / time.Millisecond)
}
