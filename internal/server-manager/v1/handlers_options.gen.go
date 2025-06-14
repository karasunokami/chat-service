// Code generated by options-gen. DO NOT EDIT.
package managerv1

import (
	fmt461e464ebed9 "fmt"

	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
	validator461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/validator"
)

type OptOptionsSetter func(o *Options)

func NewOptions(
	canReceiveProblems canReceiveProblemsUseCase,
	freeHands freeHandsUseCase,
	getChats getChatsUseCase,
	getHistory getHistoryUseCase,
	sendMessage sendMessageUseCase,
	closeChat closeChatUseCase,
	options ...OptOptionsSetter,
) Options {
	o := Options{}

	// Setting defaults from field tag (if present)

	o.canReceiveProblems = canReceiveProblems
	o.freeHands = freeHands
	o.getChats = getChats
	o.getHistory = getHistory
	o.sendMessage = sendMessage
	o.closeChat = closeChat

	for _, opt := range options {
		opt(&o)
	}
	return o
}

func (o *Options) Validate() error {
	errs := new(errors461e464ebed9.ValidationErrors)
	errs.Add(errors461e464ebed9.NewValidationError("canReceiveProblems", _validate_Options_canReceiveProblems(o)))
	errs.Add(errors461e464ebed9.NewValidationError("freeHands", _validate_Options_freeHands(o)))
	errs.Add(errors461e464ebed9.NewValidationError("getChats", _validate_Options_getChats(o)))
	errs.Add(errors461e464ebed9.NewValidationError("getHistory", _validate_Options_getHistory(o)))
	errs.Add(errors461e464ebed9.NewValidationError("sendMessage", _validate_Options_sendMessage(o)))
	errs.Add(errors461e464ebed9.NewValidationError("closeChat", _validate_Options_closeChat(o)))
	return errs.AsError()
}

func _validate_Options_canReceiveProblems(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.canReceiveProblems, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `canReceiveProblems` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_freeHands(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.freeHands, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `freeHands` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_getChats(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.getChats, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `getChats` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_getHistory(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.getHistory, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `getHistory` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_sendMessage(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.sendMessage, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `sendMessage` did not pass the test: %w", err)
	}
	return nil
}

func _validate_Options_closeChat(o *Options) error {
	if err := validator461e464ebed9.GetValidatorFor(o).Var(o.closeChat, "required"); err != nil {
		return fmt461e464ebed9.Errorf("field `closeChat` did not pass the test: %w", err)
	}
	return nil
}
