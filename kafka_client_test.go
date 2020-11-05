// +build integration

package kafka_exporter_test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/adambabik/kafka_exporter/internal/segment"
	"github.com/adambabik/kafka_exporter/pkg/types"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

const localBroker = "localhost:9092"

func TestMain(m *testing.M) {
	upDockerCompose(time.Second * 30)

	// Wait for the containers being up and running.
	done := make(chan struct{}, 1)
	go func() {
		client := segment.New(segment.Config{Broker: localBroker})
		for {
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			err := client.Healthcheck(ctx)
			if err == nil {
				close(done)
				return
			}
		}
	}()
	select {
	case <-time.After(time.Second * 60):
		downDockerCompose(time.Second * 10)
		log.Fatalf("timed out waiting for Kafka to be operational")
	case <-done:
	}

	code := m.Run()

	downDockerCompose(time.Second * 10)

	os.Exit(code)
}

func TestSegmentKafkaClient_Brokers(t *testing.T) {
	client := segment.New(segment.Config{
		Broker: localBroker,
	})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	brokers, err := client.Brokers(ctx)
	require.NoError(t, err)
	require.Len(t, brokers, 1)
}

func TestSegmentKafkaClient_Topics(t *testing.T) {
	const topic = "test-topics-1"

	client := segment.New(segment.Config{
		Broker: localBroker,
	})

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	conn, err := client.Conn(ctx)
	require.NoError(t, err)
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	tCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	topics, err := client.Topics(tCtx)
	require.NoError(t, err)
	require.Condition(t, func() (success bool) {
		for _, t := range topics {
			if t.Name == topic {
				return true
			}
		}
		return false
	})
}

func TestSegmentKafkaClient_Offsets(t *testing.T) {
	const topic = "test-offsets-1"

	client := segment.New(segment.Config{
		Broker: localBroker,
	})

	// Create a topic.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	conn, err := client.Conn(ctx)
	require.NoError(t, err)
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	// Write a message in order to bump the topic partition offset.
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{localBroker},
		Topic:   topic,
	})
	wCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = writer.WriteMessages(wCtx, kafka.Message{
		Key:   []byte("key-1"),
		Value: []byte("value-1"),
	})
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Get the offsets finally. The offset for the new topic should be bumped.
	oCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	offsets, err := client.Offsets(oCtx)
	require.NoError(t, err)
	expected := types.Offset{
		Topic: topic,
		Partitions: map[int]types.LowHighOffset{
			0: {
				Low:  0,
				High: 1,
			},
		},
	}
	require.Contains(t, offsets, expected)
}

func TestSegmentKafkaClient_ConsumerGroups(t *testing.T) {
	const (
		topic     = "test-offsets-1"
		groupName = "cg-1"
	)

	client := segment.New(segment.Config{
		Broker: localBroker,
	})

	// Create a topic.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	conn, err := client.Conn(ctx)
	require.NoError(t, err)
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	// Produce a message to the topic.
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{localBroker},
		Topic:   topic,
	})
	wCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = writer.WriteMessages(wCtx, kafka.Message{
		Key:   []byte("key-1"),
		Value: []byte("value-1"),
	})
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Create a new reader using consumer groups.
	// We need this to bump a consumer group offset
	// which should be bumped after reading a message.
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{localBroker},
		GroupID:     groupName,
		Topic:       topic,
		StartOffset: kafka.FirstOffset,
	})
	fCtx, _ := context.WithTimeout(context.Background(), time.Second*30)
	msg, err := reader.FetchMessage(fCtx)
	require.NoError(t, err)
	require.Equal(t, []byte("key-1"), msg.Key)
	err = reader.CommitMessages(fCtx, msg)
	require.NoError(t, err)
	require.NoError(t, reader.Close())

	// Check if the new consumer groups is returned.
	cgCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	groups, err := client.ConsumerGroups(cgCtx)
	require.NoError(t, err)
	require.EqualValues(t, []types.ConsumerGroup{{Name: "cg-1"}}, groups)

	// Finally, verify if the consumer group offset has changed
	// for this particular group and topic.
	require.Eventually(t, func() bool {
		ocgCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
		offsets, err := client.OffsetsForConsumerGroups(ocgCtx)
		if err != nil {
			return false
		}
		for _, o := range offsets {
			if o.Group == groupName && o.Topic == topic {
				if o.Partitions[0].High == 1 {
					return true
				}
			}
		}
		return false
	}, time.Second*30, time.Second*5)
}

func upDockerCompose(timeout time.Duration) {
	cmdStart := exec.Command("docker-compose", "up", "-d", "--build", "zookeeper", "kafka")
	go readStderr(cmdStart)
	if err := execWithTimeout(cmdStart, timeout); err != nil {
		log.Fatalf("failed to start docker-compose: %v", err)
	}
}

func downDockerCompose(timeout time.Duration) {
	cmdStop := exec.Command("docker-compose", "down")
	go readStderr(cmdStop)
	if err := execWithTimeout(cmdStop, timeout); err != nil {
		log.Fatalf("failed to stop docker-compose: %v", err)
	}
}

func execWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	done := make(chan error, 1)

	go func() { done <- cmd.Wait() }()

	select {
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("command %q timed out and failed to exit: %w", cmd.Path, err)
		}
		return fmt.Errorf("command %q timed out", cmd.Path)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("command %q failed: %w", cmd.Path, err)
		}
	}

	return nil
}

func readStderr(cmd *exec.Cmd) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("failed to get stderr from command %q: %v", cmd.Path, err)
		return
	}
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("failed reading for command %q: %v", cmd.Path, scanner.Err())
	}
}
