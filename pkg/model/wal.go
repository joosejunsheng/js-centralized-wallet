package model

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaWALMessage struct {
	TransactionUUID string `json:"transaction_uuid"`
	SourceUserId    uint64 `json:"source_user_id"`
	DestUserId      uint64 `json:"dest_user_id"`
	Amount          int64  `json:"amount"`
}

const TRANSFER_WAL = "transfer_wal"

func (m *Model) sendToKafka(ctx context.Context, producer sarama.AsyncProducer, msg []byte) error {
	message := &sarama.ProducerMessage{
		Topic: TRANSFER_WAL,
		Value: sarama.ByteEncoder(msg),
	}

	select {
	case producer.Input() <- message:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while sending to Kafka")
	}
}
