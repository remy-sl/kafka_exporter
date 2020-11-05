package kafka_exporter

import (
	"context"
	"sync"
	"testing"

	"github.com/adambabik/kafka_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector_Describe(t *testing.T) {
	ch := make(chan *prometheus.Desc, len(allMetrics)*2)
	c := new(Collector)
	c.Describe(ch)
	require.Equal(t, len(allMetrics), len(ch))
}

type mockKafkaClient struct{}

func (mockKafkaClient) Brokers(ctx context.Context) ([]types.Broker, error) {
	return []types.Broker{
		{
			ID:   1,
			Host: "127.0.0.1",
			Port: 9100,
		},
	}, nil
}

func (mockKafkaClient) Topics(ctx context.Context) ([]types.Topic, error) {
	return []types.Topic{
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
	}, nil
}

func (mockKafkaClient) Offsets(ctx context.Context) ([]types.Offset, error) {
	return []types.Offset{
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
	}, nil
}

func (mockKafkaClient) ConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error) {
	return []types.ConsumerGroup{
		{
			Name: "group-1",
		},
	}, nil
}

func (mockKafkaClient) OffsetsForConsumerGroups(ctx context.Context) ([]types.Offset, error) {
	return []types.Offset{
		{
			Topic: "topic-1",
			Group: "group-1",
			Partitions: map[int]types.LowHighOffset{
				0: {Low: 0, High: 5},
			},
		},
	}, nil
}

func TestCollector_Collect(t *testing.T) {
	ch := make(chan prometheus.Metric)

	var (
		metrics []prometheus.Metric
		wg      sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		for m := range ch {
			metrics = append(metrics, m)
		}
		wg.Done()
	}()

	c := NewCollector(&mockKafkaClient{})
	c.Collect(ch)

	close(ch)
	wg.Wait()

	assert.EqualValues(t, 1, testutil.ToFloat64(brokersTotal))
	assert.EqualValues(t, 1, testutil.ToFloat64(brokerInfo))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitions.WithLabelValues("topic-1")))
	assert.EqualValues(t, 2, testutil.ToFloat64(topicPartitions.WithLabelValues("topic-2")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionLeader.WithLabelValues("topic-1", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionLeader.WithLabelValues("topic-2", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionLeader.WithLabelValues("topic-2", "1")))
	assert.EqualValues(t, 10, testutil.ToFloat64(topicPartitionCurrentOffset.WithLabelValues("topic-1", "0")))
	assert.EqualValues(t, 10, testutil.ToFloat64(topicPartitionCurrentOffset.WithLabelValues("topic-2", "0")))
	assert.EqualValues(t, 20, testutil.ToFloat64(topicPartitionCurrentOffset.WithLabelValues("topic-2", "1")))
	assert.EqualValues(t, 0, testutil.ToFloat64(topicPartitionOldestOffset.WithLabelValues("topic-1", "0")))
	assert.EqualValues(t, 0, testutil.ToFloat64(topicPartitionOldestOffset.WithLabelValues("topic-2", "0")))
	assert.EqualValues(t, 0, testutil.ToFloat64(topicPartitionOldestOffset.WithLabelValues("topic-2", "1")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionReplicas.WithLabelValues("topic-1", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionReplicas.WithLabelValues("topic-2", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionReplicas.WithLabelValues("topic-2", "1")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionInSyncReplicas.WithLabelValues("topic-1", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionInSyncReplicas.WithLabelValues("topic-2", "0")))
	assert.EqualValues(t, 1, testutil.ToFloat64(topicPartitionInSyncReplicas.WithLabelValues("topic-2", "1")))
	assert.EqualValues(t, 1, testutil.ToFloat64(consumerGroups))
	assert.EqualValues(t, 5, testutil.ToFloat64(consumerGroupCurrentOffset.WithLabelValues("group-1", "topic-1", "0")))
}
