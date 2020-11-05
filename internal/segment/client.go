package segment

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/adambabik/kafka_exporter/pkg/types"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Config struct {
	Broker        string
	SASLUsername  string
	SASLPassword  string
	SALSMechanism string
	TLSEnabled    bool
}

func (c Config) newDialer() *kafka.Dialer {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if c.TLSEnabled {
		dialer.TLS = &tls.Config{
			RootCAs:    rootCAs(),
			MinVersion: tls.VersionTLS12,
		}
	}
	if c.SASLUsername != "" && c.SASLPassword != "" {
		dialer.SASLMechanism = plain.Mechanism{
			Username: c.SASLUsername,
			Password: c.SASLPassword,
		}
	}
	return dialer
}

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Brokers(ctx context.Context) (result []types.Broker, err error) {
	conn, err := c.dial(ctx, c.cfg.Broker)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = conn.Close()
		}
	}()

	brokers, err := conn.Brokers()
	if err != nil {
		return nil, err
	}

	result = make([]types.Broker, 0, len(brokers))
	for _, b := range brokers {
		result = append(result, types.Broker{
			ID:   b.ID,
			Host: b.Host,
			Port: b.Port,
		})
	}
	return result, nil
}

func brokerIDs(brokers []kafka.Broker) []int {
	result := make([]int, 0, len(brokers))
	for _, b := range brokers {
		result = append(result, b.ID)
	}
	return result
}

func (c *Client) Topics(ctx context.Context) (result []types.Topic, err error) {
	conn, err := c.dial(ctx, c.cfg.Broker)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = conn.Close()
		}
	}()

	// Getting partitions request must be sent to a controller node.
	// As it's unknown which node is dialed, the controller node
	// must be selected manually.
	ctrlBroker, err := conn.Controller()
	if err != nil {
		return nil, err
	}
	ctrl, err := c.dial(ctx, fmt.Sprintf("%s:%d", ctrlBroker.Host, ctrlBroker.Port))
	if err != nil {
		return nil, err
	}

	partitions, err := ctrl.ReadPartitions()
	if err != nil {
		return nil, err
	}

	topics := map[string]types.Topic{}
	for _, p := range partitions {
		t := topics[p.Topic]

		t.Name = p.Topic
		t.Partitions = append(t.Partitions, types.Partition{
			ID:       p.ID,
			Leader:   p.Leader.ID,
			Replicas: brokerIDs(p.Replicas),
			Isrs:     brokerIDs(p.Isr),
		})

		topics[p.Topic] = t
	}

	result = make([]types.Topic, 0, len(topics))
	for name, t := range topics {
		result = append(result, types.Topic{
			Name:       name,
			Partitions: t.Partitions,
		})
	}
	return result, nil
}

type offsetReq struct {
	Topic     string
	Partition int
	Leader    kafka.Broker
}

func findBroker(brokers []types.Broker, brokerID int) (types.Broker, bool) {
	for _, b := range brokers {
		if b.ID == brokerID {
			return b, true
		}
	}
	return types.Broker{}, false
}

func (c *Client) Offsets(ctx context.Context) (result []types.Offset, err error) {
	brokers, err := c.Brokers(ctx)
	if err != nil {
		return nil, err
	}

	topics, err := c.Topics(ctx)
	if err != nil {
		return nil, err
	}

	var (
		mu         sync.Mutex // mutex for resultsMap
		resultsMap = make(map[string]types.Offset)

		wg       sync.WaitGroup
		errC     = make(chan error, 1)
		requests = make(chan offsetReq, 10)
	)

	go func() {
		for oerr := range errC {
			if oerr != nil && err == nil {
				err = oerr
				return
			}
		}
	}()

	// Spawn 64 workers.
	//
	// TODO: This number is estimated based on some real-world use cases
	//  which might not be suitable for everyone.
	//  It should be configurable.
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for req := range requests {
				partitionStr := strconv.Itoa(req.Partition)

				conn, err := c.dialWithPartition(ctx, c.cfg.Broker, kafka.Partition{
					Topic:  req.Topic,
					Leader: req.Leader,
					ID:     req.Partition,
				})
				if err != nil {
					errC <- fmt.Errorf("failed to dial leader for a topic %q and partition %q: %w", req.Topic, partitionStr, err)
					return
				}

				low, high, err := conn.ReadOffsets()
				if err != nil {
					_ = conn.Close()
					errC <- fmt.Errorf("failed to read offsets for a topic %q and partition %q: %w", req.Topic, partitionStr, err)
					return
				}
				_ = conn.Close()

				mu.Lock()

				offset, ok := resultsMap[req.Topic]
				if !ok {
					offset.Topic = req.Topic
					offset.Partitions = map[int]types.LowHighOffset{}
				}

				offset.Partitions[req.Partition] = types.LowHighOffset{
					Low:  low,
					High: high,
				}

				resultsMap[req.Topic] = offset

				mu.Unlock()
			}
		}()
	}

	for _, t := range topics {
		for _, p := range t.Partitions {
			leader, ok := findBroker(brokers, p.Leader)
			if !ok {
				return nil, fmt.Errorf("failed to find a leader for a topic %q and partition %q", t, p)
			}

			requests <- offsetReq{
				Topic:     t.Name,
				Partition: p.ID,
				Leader: kafka.Broker{
					Host: leader.Host,
					Port: leader.Port,
					ID:   leader.ID,
				},
			}
		}
	}

	close(requests)
	wg.Wait()

	close(errC)

	if err != nil {
		return nil, err
	}

	for _, offset := range resultsMap {
		result = append(result, offset)
	}
	return result, nil
}

func (c *Client) ConsumerGroups(ctx context.Context) (result []types.ConsumerGroup, err error) {
	brokers, err := c.Brokers(ctx)
	if err != nil {
		return nil, err
	}

	var groups []string

	for _, broker := range brokers {
		conn, err := c.dial(ctx, fmt.Sprintf("%s:%d", broker.Host, broker.Port))
		if err != nil {
			return nil, err
		}
		items, err := conn.ListGroups()
		_ = conn.Close()
		if err != nil {
			return nil, err
		}
		groups = append(groups, items...)
	}

	result = make([]types.ConsumerGroup, 0, len(groups))
	for _, name := range groups {
		// Empty groups can be returned.
		// TODO: figure out what empty group name means.
		if name != "" {
			result = append(result, types.ConsumerGroup{Name: name})
		}
	}
	return result, nil
}

func (c *Client) OffsetsForConsumerGroups(ctx context.Context) (result []types.Offset, err error) {
	groups, err := c.ConsumerGroups(ctx)
	if err != nil {
		return nil, err
	}

	topics, err := c.Topics(ctx)
	if err != nil {
		return nil, err
	}

	topicPartitions := make(map[string][]int, len(topics))
	for _, topic := range topics {
		partitions := make([]int, 0, len(topic.Partitions))
		for _, p := range topic.Partitions {
			partitions = append(partitions, p.ID)
		}
		topicPartitions[topic.Name] = partitions
	}

	conn, err := c.dial(ctx, c.cfg.Broker)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = conn.Close()
		}
	}()

	for _, g := range groups {
		coBroker, err := conn.FindCoordinator(g.Name)
		if err != nil {
			return nil, err
		}

		coConn, err := c.dial(ctx, fmt.Sprintf("%s:%d", coBroker.Host, coBroker.Port))
		if err != nil {
			return nil, err
		}

		offsets, err := coConn.OffsetFetch(g.Name, topicPartitions)
		if err != nil {
			_ = coConn.Close()
			return nil, err
		}
		_ = coConn.Close()

		for topic, offset := range offsets {
			partitions := make(map[int]types.LowHighOffset, len(offsets))
			for pID, value := range offset {
				partitions[pID] = types.LowHighOffset{
					High: value,
				}
			}

			result = append(result, types.Offset{
				Topic:      topic,
				Group:      g.Name,
				Partitions: partitions,
			})
		}
	}

	return result, nil
}

func (c *Client) Conn(ctx context.Context) (*kafka.Conn, error) {
	return c.cfg.newDialer().DialContext(ctx, "tcp", c.cfg.Broker)
}

func (c *Client) Healthcheck(ctx context.Context) error {
	conn, err := c.Conn(ctx)
	if err != nil {
		return err
	}
	// It is required to ask for metadata to make sure
	// the node is up and running.
	_, err = conn.Brokers()
	return err
}

func (c *Client) dial(ctx context.Context, broker string) (*kafka.Conn, error) {
	return c.cfg.newDialer().DialContext(ctx, "tcp", broker)
}

func (c *Client) dialWithPartition(ctx context.Context, broker string, partition kafka.Partition) (*kafka.Conn, error) {
	return c.cfg.newDialer().DialPartition(ctx, "tcp", broker, partition)
}

func rootCAs() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	return rootCAs
}
