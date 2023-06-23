package getchats_test

import (
	"testing"

	"github.com/karasunokami/chat-service/internal/types"
	getchats "github.com/karasunokami/chat-service/internal/usecases/manager/get-chats"

	"github.com/stretchr/testify/assert"
)

func TestRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		request getchats.Request
		wantErr bool
	}{
		// Positive.
		{
			name: "cursor specified",
			request: getchats.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.NewUserID(),
			},
			wantErr: false,
		},

		// Negative.
		{
			name: "require request id",
			request: getchats.Request{
				ID:        types.RequestIDNil,
				ManagerID: types.NewUserID(),
			},
			wantErr: true,
		},
		{
			name: "require client id",
			request: getchats.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.UserIDNil,
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
