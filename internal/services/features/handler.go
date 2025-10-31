package features

import (
	"sconcur/internal/services/connection"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/features/mongodb_feature"
	"sconcur/internal/services/features/sleep_feature"
	"sconcur/internal/services/features/unknown_feature"
	"sconcur/internal/services/flows"
	"sconcur/internal/services/types"
	"sconcur/pkg/foundation/errs"
	"sync/atomic"
)

type Handler struct {
	flows   *flows.Flows
	closing atomic.Bool
}

func NewHandler() *Handler {
	return &Handler{
		flows: flows.NewFlows(),
	}
}

func (h *Handler) Handle(transport *connection.Transport, message *dto.Message) error {
	if message.Method == 1 {
		if h.closing.Load() {
			err := transport.WriteResult(
				&dto.Result{
					FlowUuid: message.FlowUuid,
					TaskKey:  message.TaskKey,
					Method:   1,
					IsError:  true,
				},
			)

			if err != nil {
				return errs.Err(err)
			}
		}

		flow := flows.NewFlow(message.FlowUuid, transport)

		h.flows.Add(message.FlowUuid, flow)

		err := flow.Run()

		h.flows.Delete(message.FlowUuid)

		return errs.Err(err)
	}

	flow, err := h.flows.Get(message.FlowUuid)

	if err != nil {
		return errs.Err(err)
	}

	flow.AddMessage(message)

	handler, err := h.detectHandler(message.Method)

	if err != nil {
		return errs.Err(err)
	}

	result := handler.Handle(flow, message)

	if result.Waitable {
		err := flow.AddResult(result)

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

func (h *Handler) Stop() {
	h.closing.Store(true)
}

func (h *Handler) detectHandler(method types.Method) (contracts.MessageHandler, error) {
	if method == 2 {
		return sleep_feature.New(), nil
	}

	if method == 3 {
		return mongodb_feature.New(), nil
	}

	return unknown_feature.New(), nil
}
