package kafka_exporter

import (
	"sync"
	"testing"

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

	c := NewCollector(newMockKafkaClient(), nopLogger)
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
