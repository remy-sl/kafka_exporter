package kafka_exporter

import (
	"context"

	"github.com/adambabik/kafka_exporter/pkg/types"
)

type mockKafkaClient struct {
	brokers       []types.Broker
	topics        []types.Topic
	offsets       []types.Offset
	groups        []types.ConsumerGroup
	groupsOffsets []types.Offset
}

func newMockKafkaClient(mods ...func(client *mockKafkaClient)) *mockKafkaClient {
	c := &mockKafkaClient{
		brokers: []types.Broker{
			{
				ID:   1,
				Host: "127.0.0.1",
				Port: 9100,
			},
		},
		topics: []types.Topic{
			{
				Name: "topic-1",
				Partitions: []types.Partition{
					{
						ID:       0,
						Leader:   1,
						Replicas: []int{1},
						Isrs:     []int{1},
					},
				},
			},
			{
				Name: "topic-2",
				Partitions: []types.Partition{
					{
						ID:       0,
						Leader:   1,
						Replicas: []int{1},
						Isrs:     []int{1},
					},
					{
						ID:       1,
						Leader:   1,
						Replicas: []int{1},
						Isrs:     []int{1},
					},
				},
			},
		},
		offsets: []types.Offset{
			{
				Topic: "topic-1",
				Partitions: map[int]types.LowHighOffset{
					0: {Low: 0, High: 10},
				},
			},
			{
				Topic: "topic-2",
				Partitions: map[int]types.LowHighOffset{
					0: {Low: 0, High: 10},
					1: {Low: 0, High: 20},
				},
			},
		},
		groups: []types.ConsumerGroup{
			{
				Name: "group-1",
			},
		},
		groupsOffsets: []types.Offset{
			{
				Topic: "topic-1",
				Group: "group-1",
				Partitions: map[int]types.LowHighOffset{
					0: {Low: 0, High: 5},
				},
			},
		},
	}

	for _, mod := range mods {
		mod(c)
	}

	return c
}

func (c mockKafkaClient) Brokers(context.Context) ([]types.Broker, error) {
	return c.brokers, nil
}

func (c mockKafkaClient) Topics(context.Context) ([]types.Topic, error) {
	return c.topics, nil
}

func (c mockKafkaClient) Offsets(context.Context) ([]types.Offset, error) {
	return c.offsets, nil
}

func (c mockKafkaClient) ConsumerGroups(context.Context) ([]types.ConsumerGroup, error) {
	return c.groups, nil
}

func (c mockKafkaClient) OffsetsForConsumerGroups(context.Context) ([]types.Offset, error) {
	return c.groupsOffsets, nil
}
