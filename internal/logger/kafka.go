package logger

import (
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ kafka.Logger = (*KafkaAdapted)(nil)

type KafkaAdapted struct {
	atom   zap.AtomicLevel
	logger *zap.Logger
}

func NewKafkaAdapted() *KafkaAdapted {
	k := &KafkaAdapted{}

	k.init()

	return k
}

func (k *KafkaAdapted) Printf(tpl string, args ...interface{}) {
	k.logger.Info(fmt.Sprintf(tpl, args...))
}

func (k *KafkaAdapted) WithServiceName(serviceName string) *KafkaAdapted {
	k.logger = k.logger.Named(serviceName)

	return k
}

func (k *KafkaAdapted) ForErrors() *KafkaAdapted {
	k.atom.SetLevel(zapcore.ErrorLevel)

	return k
}

func (k *KafkaAdapted) init() {
	k.atom = zap.NewAtomicLevel()

	encoder := createEncoder(false)
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), k.atom),
	}

	k.logger = zap.New(zapcore.NewTee(cores...))

	err := k.logger.Sync()
	if err != nil {
		zap.L().Error("sync kafka logger", zap.Error(err))
	}
}
