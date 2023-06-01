package freehands_test

import (
	"testing"

	"github.com/karasunokami/chat-service/internal/types"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"

	"github.com/stretchr/testify/assert"
)

func TestRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		request freehands.Request
		wantErr bool
	}{
		{
			name: "valid",
			request: freehands.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.NewUserID(),
			},
			wantErr: false,
		},
		{
			name: "empty id",
			request: freehands.Request{
				ID:        types.RequestIDNil,
				ManagerID: types.NewUserID(),
			},
			wantErr: true,
		},
		{
			name: "empty user id",
			request: freehands.Request{
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
