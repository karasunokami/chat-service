package closechat_test

import (
	"testing"

	"github.com/karasunokami/chat-service/internal/types"
	closechat "github.com/karasunokami/chat-service/internal/usecases/manager/close-chat"

	"github.com/stretchr/testify/assert"
)

func TestRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		request closechat.Request
		wantErr bool
	}{
		{
			name: "valid",
			request: closechat.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.NewUserID(),
				ChatID:    types.NewChatID(),
			},
			wantErr: false,
		},
		{
			name: "empty id",
			request: closechat.Request{
				ID:        types.RequestIDNil,
				ManagerID: types.NewUserID(),
				ChatID:    types.NewChatID(),
			},
			wantErr: true,
		},
		{
			name: "empty user id",
			request: closechat.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.UserIDNil,
				ChatID:    types.NewChatID(),
			},
			wantErr: true,
		},
		{
			name: "empty chat id",
			request: closechat.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.NewUserID(),
				ChatID:    types.ChatIDNil,
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
