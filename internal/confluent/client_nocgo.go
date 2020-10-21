// +build !cgo

package confluent

import (
	"context"

	"github.com/adambabik/kafka_exporter/pkg/types"
)

type ConfigMap map[string]interface{}

type Client struct{}

func New(ConfigMap) *Client {
	return &Client{}
}

func (c *Client) Brokers(ctx context.Context) (result []types.Broker, err error) {
	return
}

func (c *Client) Topics(ctx context.Context) (result []types.Topic, err error) {
	return
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
