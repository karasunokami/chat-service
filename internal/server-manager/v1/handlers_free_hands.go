package managerv1

import (
	"fmt"
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostGetFreeHandsBtnAvailability(eCtx echo.Context, params PostGetFreeHandsBtnAvailabilityParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)

	r := canreceiveproblems.Request{
		ID:        params.XRequestID,
		ManagerID: clientID,
	}

	resp, err := h.canReceiveProblems.Handle(ctx, r)
	if err != nil {
		return fmt.Errorf("can receive problems handle, err=%v", err)
	}

	return eCtx.JSON(http.StatusOK, GetFreeHandsBtnAvailabilityResponse{
		Data: &ManagerAvailability{
			Available: resp.Result,
		},
	})
}
