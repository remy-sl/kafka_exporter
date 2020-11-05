package kafka_exporter

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/adambabik/kafka_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	client KafkaClient
}

var _ prometheus.Collector = (*Collector)(nil) // ensure Collector implements prometheus.Collector interface

func NewCollector(client KafkaClient) *Collector {
	return &Collector{client: client}
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
				log.Printf("failed to resolve %q: %v", b.Host, err)
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
		log.Printf("failed to get brokers: %v", err)
		return nil
	}
	log.Printf("got brokers: %+v", brokers)
	return brokers
}

func (e *Collector) topics() []types.Topic {
	topics, err := e.client.Topics(context.Background())
	if err != nil {
		log.Printf("failed to get topics: %v", err)
		return nil
	}
	log.Printf("got topics: %+v", topics)
	return topics
}

func (e *Collector) offsets() []types.Offset {
	offsets, err := e.client.Offsets(context.Background())
	if err != nil {
		log.Printf("failed to get offsets: %v", err)
		return nil
	}
	log.Printf("got offsets: %+v", offsets)
	return offsets
}

func (e *Collector) consumerGroups() []types.ConsumerGroup {
	groups, err := e.client.ConsumerGroups(context.Background())
	if err != nil {
		log.Printf("failed to get consumer groups: %v", err)
		return nil
	}
	log.Printf("got consumer groups: %+v", groups)
	return groups
}

func (e *Collector) offsetsForConsumerGroups() []types.Offset {
	offsets, err := e.client.OffsetsForConsumerGroups(context.Background())
	if err != nil {
		log.Printf("failed to get offsets for consumer groups: %v", err)
		return nil
	}
	log.Printf("got offsets for consumer groups: %+v", offsets)
	return offsets
}
