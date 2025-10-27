package unknown_feature

import (
	"context"
	"fmt"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
	Seconds int
}

func New() *Feature {
	return &Feature{}
}

func (s *Feature) Handle(_ context.Context, message *dto.Message) *dto.Result {
	return &dto.Result{
		FlowUuid: message.FlowUuid,
		Method:   message.Method,
		TaskKey:  message.TaskKey,
		Waitable: true,
		IsError:  true,
		Payload: fmt.Sprintf(
			"Unknown method: %d",
			message.Method,
		),
	}
}
