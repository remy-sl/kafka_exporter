package main

import (
	"github.com/adambabik/kafka_exporter/internal/confluent"
	"github.com/adambabik/kafka_exporter/internal/segment"
	"log"
	"os"

	"github.com/adambabik/kafka_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "kafka_exporter",
		Usage: "collects high-level metrics from a Kafka cluster",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Aliases: []string{"a"},
				Value:   "127.0.0.1:9150",
				Usage:   "bind address",
			},
			&cli.StringFlag{
				Name:  "client",
				Value: "segment",
				Usage: "a name of the Kafka client to use (possible choices are: confluent, segment)",
			},
			&cli.StringFlag{
				Name:    "broker",
				Aliases: []string{"b"},
				Value:   "localhost:9092",
				Usage:   "a Kafka broker address",
			},
			&cli.StringFlag{
				Name:    "sasl.username",
				EnvVars: []string{"KAFKA_SASL_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "sasl.password",
				EnvVars: []string{"KAFKA_SASL_PASSWORD"},
			},
			&cli.StringFlag{
				Name: "sasl.mechanisms", // typical values: PLAIN
			},
			&cli.StringFlag{
				Name:  "security.protocol", // typical values: SASL_SSL
				Usage: "a security protocol; only for confluent client",
			},
			&cli.BoolFlag{
				Name:  "tls",
				Usage: "true enables TLS; only for segment client",
			},
			&cli.StringFlag{
				Name:  "ssl.ca.location", // on macOS with OpenSSL use something like "/usr/local/etc/openssl@1.1/cert.pem"
				Usage: "location of a SSL certificate authority; only for confluent client",
			},
		},
		Action: func(c *cli.Context) (err error) {
			var client kafka_exporter.KafkaClient

			switch c.String("client") {
			case "confluent":
				if !confluentClientSupported {
					panic("confluent client is not supported")
				}

				cfg := confluent.ConfigMap{
					"bootstrap.servers":       c.String("broker"),
					"api.version.request":     "true",
					"api.version.fallback.ms": 0,
				}
				if val := c.String("sasl.mechanisms"); val != "" {
					cfg["sasl.mechanisms"] = val
				}
				if val := c.String("security.protocol"); val != "" {
					cfg["security.protocol"] = val
				}
				if val := c.String("sasl.username"); val != "" {
					cfg["sasl.username"] = val
				}
				if val := c.String("sasl.password"); val != "" {
					cfg["sasl.password"] = val
				}
				if val := c.String("ssl.ca.location"); val != "" {
					cfg["ssl.ca.location"] = val
				}
				client = confluent.New(cfg)
			case "segment":
				client = segment.New(segment.Config{
					Broker:        c.String("broker"),
					SASLUsername:  c.String("sasl.username"),
					SASLPassword:  c.String("sasl.password"),
					SALSMechanism: c.String("sasl.mechanisms"),
					TLSEnabled:    c.Bool("tls"),
				})
			}

			exporter := kafka_exporter.NewCollector(client)

			if err := prometheus.Register(exporter); err != nil {
				return err
			}

			address := c.String("address")
			log.Printf("listening on address %s", address)
			return kafka_exporter.ListenAndServe("/metrics", address)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("kafka_exporter failed: %v", err)
	}
}
