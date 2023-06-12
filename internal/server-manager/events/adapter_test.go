package managerevents_test

import (
	"encoding/json"
	"testing"

	managerevents "github.com/karasunokami/chat-service/internal/server-manager/events"
	eventstream "github.com/karasunokami/chat-service/internal/services/event-stream"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Adapt(t *testing.T) {
	cases := []struct {
		name    string
		ev      eventstream.Event
		expJSON string
	}{
		{
			name: "smoke",
			ev: eventstream.NewNewChatEvent(
				true,
				types.MustParse[types.EventID]("d0ffbd36-bc30-11ed-8286-461e464ebed8"),
				types.MustParse[types.RequestID]("cee5f290-bc30-11ed-b7fe-461e464ebed8"),
				types.MustParse[types.ChatID]("cb36a888-bc30-11ed-b843-461e464ebed8"),
				types.MustParse[types.UserID]("4d55ddf0-3216-48a3-b5f8-3b6bb72980ec"),
			),
			expJSON: `{
				"eventId": "d0ffbd36-bc30-11ed-8286-461e464ebed8",
				"eventType": "NewChatEvent",
				"chatId": "cb36a888-bc30-11ed-b843-461e464ebed8",
				"requestId": "cee5f290-bc30-11ed-b7fe-461e464ebed8",
				"clientId": "4d55ddf0-3216-48a3-b5f8-3b6bb72980ec",
				"canTakeMoreProblems": true
			}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			adapted, err := managerevents.Adapter{}.Adapt(tt.ev)
			require.NoError(t, err)

			raw, err := json.Marshal(adapted)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expJSON, string(raw))
		})
	}
}
