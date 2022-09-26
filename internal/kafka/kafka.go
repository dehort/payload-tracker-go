package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	config "github.com/redhatinsights/payload-tracker-go/internal/config"
	"github.com/redhatinsights/payload-tracker-go/internal/endpoints"
	l "github.com/redhatinsights/payload-tracker-go/internal/logging"
)

// NewConsumer Creates brand new consumer instance based on topic
func NewConsumer(ctx context.Context, config *config.TrackerConfig, topic string) (*kafka.Consumer, error) {
	var configMap kafka.ConfigMap

	if config.KafkaConfig.SASLMechanism != "" {
		configMap = kafka.ConfigMap{
			"bootstrap.servers":        config.KafkaConfig.KafkaBootstrapServers,
			"group.id":                 config.KafkaConfig.KafkaGroupID,
			"security.protocol":        config.KafkaConfig.Protocol,
			"sasl.mechanism":           config.KafkaConfig.SASLMechanism,
			"ssl.ca.location":          config.KafkaConfig.KafkaCA,
			"sasl.username":            config.KafkaConfig.KafkaUsername,
			"sasl.password":            config.KafkaConfig.KafkaPassword,
			"go.logs.channel.enable":   true,
			"allow.auto.create.topics": true,
		}
	} else {
		configMap = kafka.ConfigMap{
			"bootstrap.servers":        config.KafkaConfig.KafkaBootstrapServers,
			"group.id":                 config.KafkaConfig.KafkaGroupID,
			"auto.offset.reset":        config.KafkaConfig.KafkaAutoOffsetReset,
			"auto.commit.interval.ms":  config.KafkaConfig.KafkaAutoCommitInterval,
			"go.logs.channel.enable":   true,
			"allow.auto.create.topics": true,
		}
	}

	consumer, err := kafka.NewConsumer(&configMap)

	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)

	if err != nil {
		return nil, err
	}

	l.Log.Info("Connected to Kafka")

	return consumer, nil
}

// NewConsumerEventLoop creates a new consumer event loop based on the information passed with it
func NewConsumerEventLoop(
	ctx context.Context,
	consumer *kafka.Consumer,
	messageHandler MessageHandler,
) {

	i := 0
	for {
		select {
		case <-ctx.Done():
			l.Log.Infof("Received shutdown signal.  Ending consume loop.")
			return
		default:

			event := consumer.Poll(100)
			if event == nil {
				continue
			}

			i = i + 1
			//fmt.Println("event:", event)
			fmt.Println("i:", i)

			switch e := event.(type) {
			case *kafka.Message:
				endpoints.IncConsumedMessages()
				messageHandler.onMessage(ctx, e)
			case kafka.Error:
				endpoints.IncConsumeErrors()
				l.Log.Errorf("Consumer error: %v (%v)\n", e.Code(), e)
			default:
				l.Log.Infof("Ignored %v\n", e)
			}

		}
	}
}
