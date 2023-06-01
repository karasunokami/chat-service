package managerv1

import (
	"net/http"

	"github.com/karasunokami/chat-service/internal/middlewares"
	canreceiveproblems "github.com/karasunokami/chat-service/internal/usecases/manager/can-receive-problems"
	freehands "github.com/karasunokami/chat-service/internal/usecases/manager/free-hands"

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
		return newHandleError(err, getErrorCode(err))
	}

	return eCtx.JSON(http.StatusOK, GetFreeHandsBtnAvailabilityResponse{
		Data: &ManagerAvailability{
			Available: resp.Result,
		},
	})
}

func (h Handlers) PostFreeHands(eCtx echo.Context, params PostFreeHandsParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)

	r := freehands.Request{
		ID:        params.XRequestID,
		ManagerID: clientID,
	}

	err := h.freeHands.Handle(ctx, r)
	if err != nil {
		return newHandleError(err, getErrorCode(err))
	}

	return eCtx.JSON(http.StatusOK, FreeHandsResponse{})
}
