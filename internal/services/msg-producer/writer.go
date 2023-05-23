package msgproducer

import (
	"github.com/karasunokami/chat-service/internal/logger"

	"github.com/segmentio/kafka-go"
)

const serviceName = "msg-producer"

func NewKafkaWriter(brokers []string, topic string, batchSize int) KafkaWriter {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     messagesBalancer{},
		BatchSize:    batchSize,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Logger:       logger.NewKafkaAdapted().WithServiceName(serviceName),
		ErrorLogger:  logger.NewKafkaAdapted().WithServiceName(serviceName).ForErrors(),
	}
}

type messagesBalancer struct{}

func (s messagesBalancer) Balance(msg kafka.Message, partitions ...int) (partition int) {
	return partitions[(len(partitions)-1)%int(msg.Key[0])]
}
