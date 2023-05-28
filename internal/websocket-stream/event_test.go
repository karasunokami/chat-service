package websocketstream_test

import (
	"bytes"
	"testing"

	websocketstream "github.com/karasunokami/chat-service/internal/websocket-stream"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONEventWriter_Smoke(t *testing.T) {
	wr := websocketstream.JSONEventWriter{}
	out := bytes.NewBuffer(nil)
	err := wr.Write(struct{ Name string }{Name: "John"}, out)
	require.NoError(t, err)
	assert.JSONEq(t, `{"Name":"John"}`, out.String())
}
