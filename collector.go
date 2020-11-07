package kafka_exporter

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/adambabik/kafka_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Collector struct {
	client KafkaClient
	logger *zap.Logger
}

var _ prometheus.Collector = (*Collector)(nil) // ensure Collector implements prometheus.Collector interface

var nopLogger = zap.NewNop()

func NewCollector(client KafkaClient, logger *zap.Logger) *Collector {
	if logger == nil {
		logger = nopLogger
	}
	return &Collector{client: client, logger: logger}
}

func (e *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range allMetrics {
		m.Describe(ch)
	}
}

func (e *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		collectTime.With(nil).Observe(time.Since(start).Seconds())
		collectTime.Collect(ch)
	}()

	var wg sync.WaitGroup

	// Collect information about Kafka brokers.
	wg.Add(1)
	go func() {
		defer wg.Done()

		brokers := e.brokers()
		brokersTotal.With(nil).Set(float64(len(brokers)))
		brokersTotal.Collect(ch)

		for _, b := range brokers {
			addrs, err := net.LookupHost(b.Host)
			if err != nil {
				e.logger.Error("failed to resolve broker host", zap.Any("broker", b), zap.Error(err))
			}
			brokerInfo.WithLabelValues(strconv.Itoa(b.ID), b.Host, strconv.Itoa(b.Port), addrs[0]).Set(1)
		}
		brokerInfo.Collect(ch)
	}()

	// Collect information about consumer groups.
	wg.Add(1)
	go func() {
		defer wg.Done()

		groups := e.consumerGroups()
		consumerGroups.With(nil).Set(float64(len(groups)))
		consumerGroups.Collect(ch)
	}()

	// Collect information about offsets for consumer groups.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, o := range e.offsetsForConsumerGroups() {
			for pID, p := range o.Partitions {
				consumerGroupCurrentOffset.WithLabelValues(o.Group, o.Topic, strconv.Itoa(pID)).Set(float64(p.High))
			}
		}
		consumerGroupCurrentOffset.Collect(ch)
	}()

	// Collect information about topics.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, t := range e.topics() {
			topicPartitions.WithLabelValues(t.Name).Set(float64(len(t.Partitions)))

			for _, p := range t.Partitions {
				topicPartitionReplicas.WithLabelValues(t.Name, strconv.Itoa(p.ID)).Set(float64(len(p.Replicas)))
				topicPartitionInSyncReplicas.WithLabelValues(t.Name, strconv.Itoa(p.ID)).Set(float64(len(p.Isrs)))
				topicPartitionLeader.WithLabelValues(t.Name, strconv.Itoa(p.ID)).Set(float64(p.Leader))
			}
		}

		topicPartitions.Collect(ch)
		topicPartitionReplicas.Collect(ch)
		topicPartitionInSyncReplicas.Collect(ch)
		topicPartitionLeader.Collect(ch)
	}()

	// Collect information about offsets.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, o := range e.offsets() {
			for pID, p := range o.Partitions {
				topicPartitionCurrentOffset.WithLabelValues(o.Topic, strconv.Itoa(pID)).Set(float64(p.High))
				topicPartitionOldestOffset.WithLabelValues(o.Topic, strconv.Itoa(pID)).Set(float64(p.Low))
			}
		}

		topicPartitionCurrentOffset.Collect(ch)
		topicPartitionOldestOffset.Collect(ch)
	}()

	wg.Wait()
}

func (e *Collector) brokers() []types.Broker {
	brokers, err := e.client.Brokers(context.Background())
	if err != nil {
		e.logger.Error("failed to get brokers", zap.Error(err))
		return nil
	}
	e.logger.Debug("got brokers", zap.Any("brokers", brokers))
	return brokers
}

func (e *Collector) topics() []types.Topic {
	topics, err := e.client.Topics(context.Background())
	if err != nil {
		e.logger.Error("failed to get topics", zap.Error(err))
		return nil
	}
	e.logger.Debug("got topics", zap.Any("topics", topics))
	return topics
}

func (e *Collector) offsets() []types.Offset {
	offsets, err := e.client.Offsets(context.Background())
	if err != nil {
		e.logger.Error("failed to get offsets", zap.Error(err))
		return nil
	}
	e.logger.Debug("got offsets", zap.Any("offsets", offsets))
	return offsets
}

func (e *Collector) consumerGroups() []types.ConsumerGroup {
	groups, err := e.client.ConsumerGroups(context.Background())
	if err != nil {
		e.logger.Error("failed to get consumer groups", zap.Error(err))
		return nil
	}
	e.logger.Debug("got consumer groups", zap.Any("groups", groups))
	return groups
}

func (e *Collector) offsetsForConsumerGroups() []types.Offset {
	offsets, err := e.client.OffsetsForConsumerGroups(context.Background())
	if err != nil {
		e.logger.Error("failed to get offsets for consumer groups", zap.Error(err))
		return nil
	}
	e.logger.Debug("got offsets for consumer groups", zap.Any("offsets", offsets))
	return offsets
}
