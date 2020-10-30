# kafka_exporter

Kafka high-level Prometheus metrics exporter. For low-level metrics take a look at [jmx_exporter](https://github.com/prometheus/jmx_exporter).

This exporter supports multiple, namely [kafka-go](https://github.com/segmentio/kafka-go) and [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go). kafka-go has more features and does not require CGO, hence, it's recommended. **All releases and Docker images do not support Confluent client. If you want to use it, please build from the source.**

## Metrics

Below you can find a list of supported metrics. Metrics support is different per client and subject to change in the future.

### Brokers

| Name                  | Description                       | Clients                      |
| --------------------- | --------------------------------- | ---------------------------- |
| `kafka_brokers_total` | A number of brokers in a cluster. | ✅ segment ✅ confluent     |
| `kafka_broker_info`   | Information about a broker.       | ✅ segment ✅ confluent     |

### Topics

| Name                                           | Description                                             | Clients                      |
| ---------------------------------------------- | ------------------------------------------------------- | ---------------------------- |
| `kafka_topic_partitions_total`                 | A number of partitions for a topic.                     | ✅ segment ✅ confluent     |
| `kafka_topic_partition_leader`                 | Leader's broker Name of a topic at a partition.         | ✅ segment ✅ confluent     |
| `kafka_topic_partition_current_offset`         | A current offset of a topic at a partition.             | ✅ segment ❌ confluent     |
| `kafka_topic_partition_oldest_offset`          | An oldest offset of a topic at a partition.             | ✅ segment ❌ confluent     |
| `kafka_topic_partition_replicas_total`         | A number of replicas of a topic at a partition.         | ✅ segment ✅ confluent     |
| `kafka_topic_partition_in_sync_replicas_total` | A number of in-sync replicas of a topic at a partition. | ✅ segment ✅ confluent     |

### Consumer groups

| Name                                           | Description                                                     | Clients                      |
| ---------------------------------------------- | --------------------------------------------------------------- | ---------------------------- |
| `kafka_consumer_groups_total`                  | A number of consumer groups.                                    | ✅ segment ❌ confluent     |
| `kafka_consumer_group_current_offset`          | A current offset of a consumer group for a topic and partition. | ✅ segment ❌ confluent     |

## Usage

Usually, you will run kafka_exporter as a docker container and connect it to your Kafka cluster.

The public Docker image is available on [Docker Hub](https://hub.docker.com/r/adambabik/kafka_exporter).

Alternatively, you can use a binary directly. Check out [Building > Binary](#binary) and run `kafka_exporter -h` or `go run ./cmd/kafka_exporter -h` in order to display all available flags.

In the case you want to test it locally, we recommend using `docker-compose up -d --build` and opening `http://localhost:9150` in the browser.

## Building

### Docker

```shell script
$ docker build -t kafka_exporter .
```

### Binary

```shell script
$ go build -o kafka_exporter ./cmd/kafka_exporter
```

## Related projects

This project was inspired by [github.com/danielqsj/kafka_exporter](https://github.com/danielqsj/kafka_exporter). The main difference is that it supports many Kafka clients. 

## License

MIT License