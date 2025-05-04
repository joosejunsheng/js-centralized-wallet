package server

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type ConsumerHandler struct {
	ready chan bool
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		fmt.Printf("Kafka - Received message: key = %s, value = %s, topic = %s, partition = %d, offset = %d\n",
			string(message.Key), string(message.Value), message.Topic, message.Partition, message.Offset)

		// TODO: Update Redis

		session.MarkMessage(message, "")
	}
	return nil
}

func (s *Server) StartKafkaRedisConsumer(ctx context.Context) error {
	brokers := []string{"kafka:9092"}
	groupID := "redis-cache-rebuilder"
	topic := "transfer-wal"

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer := &ConsumerHandler{
		ready: make(chan bool),
	}

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return fmt.Errorf("error creating consumer group client: %w", err)
	}
	defer client.Close()

	go func() {
		for {
			if err := client.Consume(ctx, []string{topic}, consumer); err != nil {
				log.Fatalf("error from consumer: %v", err)
			}
			// Recreate `ready` for next Consume loop
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Wait till setup is done
	log.Println("Kafka consumer ready and listening...")
	<-ctx.Done()
	return nil
}
