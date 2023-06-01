package logger

import (
	"go.uber.org/zap"
)

type KafkaAdapted struct {
	forErrors bool
	z         *zap.Logger
}

func NewKafkaAdapted() *KafkaAdapted {
	return &KafkaAdapted{
		z: zap.L(),
	}
}

func (k *KafkaAdapted) ForErrors() *KafkaAdapted {
	k.forErrors = true
	return k
}

func (k *KafkaAdapted) WithServiceName(n string) *KafkaAdapted {
	k.z = k.z.Named(n)
	return k
}

func (k *KafkaAdapted) Printf(s string, args ...interface{}) {
	if k.forErrors {
		k.z.Sugar().Errorf(s, args...)
	} else {
		k.z.Sugar().Debugf(s, args...)
	}
}
