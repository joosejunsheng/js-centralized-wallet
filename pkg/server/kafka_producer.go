package server

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func (s *Server) StartKafkaProducer() error {

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	brokers := []string{"kafka:9092"}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	defer producer.AsyncClose()

	s.KafkaProducer = producer

	// Monitor for successes and errors asynchronously
	go func() {
		for {
			select {
			case success := <-producer.Successes():
				fmt.Printf("Kafka - Message sent to topic %s: %s\n", success.Topic, success.Value)
			case err := <-producer.Errors():
				fmt.Printf("Failed to send message: %v\n", err)
			}
		}
	}()

	select {}
}
