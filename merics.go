package kafka_exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "kafka"

var (
	collectTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "kafka_exporter",
		Name:      "collect_time_seconds",
		Help:      "Time needed to collect all metrics.",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 15, 30, 60},
	}, nil)

	brokersTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "brokers_total",
		Help:      "A number of brokers in a cluster.",
	}, nil)

	topicPartitions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partitions_total",
		Help:      "A number of partitions for a topic.",
	}, []string{"topic"})

	topicPartitionLeader = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partition_leader",
		Help:      "Leader's broker Name of a topic at a partition.",
	}, []string{"topic", "partition"})

	topicPartitionCurrentOffset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partition_current_offset",
		Help:      "A current offset of a topic at a partition.",
	}, []string{"topic", "partition"})

	topicPartitionOldestOffset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partition_oldest_offset",
		Help:      "An oldest offset of a topic at a partition.",
	}, []string{"topic", "partition"})

	topicPartitionReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partition_replicas_total",
		Help:      "A number of replicas of a topic at a partition.",
	}, []string{"topic", "partition"})

	topicPartitionInSyncReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "topic_partition_in_sync_replicas_total",
		Help:      "A number of in-sync replicas of a topic at a partition.",
	}, []string{"topic", "partition"})

	consumerGroups = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "consumer_groups_total",
		Help:      "A number of consumer groups.",
	}, nil)

	consumerGroupCurrentOffset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "consumer_group_current_offset",
		Help:      "A current offset of a consumer group for a topic and partition.",
	}, []string{"group_name", "topic", "partition"})
)

var allMetrics = []prometheus.Collector{
	collectTime,

	brokersTotal,
	topicPartitions,
	topicPartitionLeader,
	topicPartitionCurrentOffset,
	topicPartitionOldestOffset,
	topicPartitionReplicas,
	topicPartitionInSyncReplicas,
	consumerGroups,
	consumerGroupCurrentOffset,
}
