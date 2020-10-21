module github.com/adambabik/kafka_exporter

go 1.15

require (
	github.com/confluentinc/confluent-kafka-go v1.4.2 // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/segmentio/kafka-go v0.4.6
	github.com/urfave/cli/v2 v2.2.0
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.4.2
)

replace github.com/segmentio/kafka-go => github.com/adambabik/kafka-go v0.3.11-0.20201021171222-e215a9a96d2a
