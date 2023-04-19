package keycloakclient

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

//go:generate options-gen -out-filename=client_options.gen.go -from-struct=Options
type Options struct {
	basePath     string `option:"mandatory" validate:"url"`
	realm        string `option:"mandatory"`
	clientID     string `option:"mandatory"`
	clientSecret string `option:"mandatory" validate:"alphanum"`

	debugMode bool
}

// Client is a tiny client to the KeyCloak realm operations. UMA configuration:
// http://localhost:3010/realms/Bank/.well-known/uma2-configuration
type Client struct {
	cli          *resty.Client
	basePath     string
	realm        string
	clientID     string
	clientSecret string
}

func New(opts Options) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %v", err)
	}

	cli := resty.New()
	cli.SetDebug(opts.debugMode)
	cli.SetBaseURL(opts.basePath)

	return &Client{
		cli:          cli,
		realm:        opts.realm,
		clientID:     opts.clientID,
		clientSecret: opts.clientSecret,
	}, nil
}
