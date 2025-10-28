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
	"sconcur/internal/services/types"
)

type Handler struct {
	flows *flows.Flows
}

func NewHandler() *Handler {
	return &Handler{
		flows: flows.GetFlows(),
	}
}

func (h *Handler) Handle(ctx context.Context, transport *connection.Transport, message *dto.Message) error {
	h.flows.AddMessage(message)

	result := h.detectHandler(transport, message.Method).Handle(ctx, message)

	if result.Waitable {
		h.flows.AddResult(result)
	}

	if !result.HasNext { // TODO

	}

	return nil
}

func (h *Handler) GetFlowsCount() int {
	return h.flows.GetCount()
}

func (h *Handler) detectHandler(transport *connection.Transport, method types.Method) contracts.MessageHandler {
	if method == 1 {
		return read_feature.New(h.flows, transport)
	}

	if method == 2 {
		return sleep_feature.New()
	}

	return unknown_feature.New()
}
