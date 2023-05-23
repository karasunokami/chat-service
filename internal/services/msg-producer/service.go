package msgproducer

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/segmentio/kafka-go"
)

type KafkaWriter interface {
	io.Closer
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

//go:generate options-gen -out-filename=service_options.gen.go -from-struct=Options
type Options struct {
	wr           KafkaWriter `option:"mandatory" validate:"required"`
	encryptKey   string      `validate:"omitempty,hexadecimal"`
	nonceFactory func(size int) ([]byte, error)
}

type Service struct {
	wr           KafkaWriter
	cipher       cipher.AEAD
	nonceFactory func(size int) ([]byte, error)
}

func New(opts Options) (*Service, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate options, err=%v", err)
	}

	if opts.nonceFactory == nil {
		opts.nonceFactory = defaultNonceFactory
	}

	s := &Service{
		wr:           opts.wr,
		nonceFactory: opts.nonceFactory,
	}

	if opts.encryptKey != "" {
		s.cipher, err = initializeCipher(opts.encryptKey)
		if err != nil {
			return nil, fmt.Errorf("initialize aead cipher, err=%v", err)
		}
	}

	return s, nil
}

func defaultNonceFactory(size int) (nonce []byte, err error) {
	nonce = make([]byte, size)

	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("rand read, err=%v", err)
	}

	return
}

func initializeCipher(key string) (cipher.AEAD, error) {
	dk, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("decode key, err=%v", err)
	}

	blockCipher, err := aes.NewCipher(dk)
	if err != nil {
		return nil, fmt.Errorf("create block cipher, err=%v", err)
	}

	aead, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("create gcm aead, err=%v", err)
	}

	return aead, nil
}
