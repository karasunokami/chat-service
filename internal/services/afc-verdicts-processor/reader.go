package afcverdictsprocessor

import (
	"context"
	"io"

	"github.com/karasunokami/chat-service/internal/logger"

	"github.com/segmentio/kafka-go"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/reader_mock.gen.go -package=afcverdictsprocessormocks

type KafkaReaderFactory func(brokers []string, groupID string, topic string) KafkaReader

type KafkaReader interface {
	io.Closer
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

func NewKafkaReader(brokers []string, groupID string, topic string) KafkaReader {
	return kafka.NewReader(kafka.ReaderConfig{
		WatchPartitionChanges: true,
		Brokers:               brokers,
		GroupID:               groupID,
		Topic:                 topic,
		Logger:                logger.NewKafkaAdapted().WithServiceName(groupID),
		ErrorLogger:           logger.NewKafkaAdapted().WithServiceName(groupID).ForErrors(),
		StartOffset:           kafka.FirstOffset,
	})
}
