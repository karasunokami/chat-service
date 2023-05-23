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

	// To keep the example deterministic, disable timestamps in the output.
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""

	k.logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		k.atom,
	))

	err := k.logger.Sync()
	if err != nil {
		zap.L().Error("sync kafka logger", zap.Error(err))
	}
}
