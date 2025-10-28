package features

import (
	"context"
	"sconcur/internal/services/connection"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/features/read_feature"
	"sconcur/internal/services/features/sleep_feature"
	"sconcur/internal/services/features/unknown_feature"
	"sconcur/internal/services/flows"
	"sconcur/pkg/foundation/errs"
)

type Handler struct {
	flows *flows.Flows
}

func NewHandler() *Handler {
	return &Handler{
		flows: flows.NewFlows(),
	}
}

func (h *Handler) Handle(ctx context.Context, transport *connection.Transport, message *dto.Message) error {
	handler, err := h.prepareHandler(
		transport,
		message,
	)

	if err != nil {
		return errs.Err(err)
	}

	result := handler.Handle(ctx, message)

	if result.Waitable {
		err := h.flows.AddResult(result)

		if err != nil {
			return errs.Err(err)
		}
	}

	if !result.HasNext { // TODO

	}

	return nil
}

func (h *Handler) GetFlowsCount() int {
	return h.flows.GetCount()
}

func (h *Handler) prepareHandler(
	transport *connection.Transport,
	message *dto.Message,
) (contracts.MessageHandler, error) {
	if message.Method == 1 {
		flow := flows.NewFlow()

		h.flows.Add(message.FlowUuid, flow)

		return read_feature.New(flow, transport), nil
	}

	flow, err := h.flows.Get(message.FlowUuid)

	if err != nil {
		return nil, errs.Err(err)
	}

	flow.AddMessage(message)

	if message.Method == 2 {
		return sleep_feature.New(flow), nil
	}

	return unknown_feature.New(), nil
}
