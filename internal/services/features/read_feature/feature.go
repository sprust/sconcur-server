package read_feature

import (
	"context"
	"log/slog"
	"sconcur/internal/services/connection"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/flows"
	"sconcur/internal/services/logging"
	"sconcur/pkg/foundation/errs"
	"sync/atomic"
	"time"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
	flows     *flows.Flows
	transport *connection.Transport
	closing   atomic.Bool
}

func New(flows *flows.Flows, transport *connection.Transport) *Feature {
	return &Feature{
		flows:     flows,
		transport: transport,
	}
}

func (f *Feature) Handle(ctx context.Context, message *dto.Message) *dto.Result {
	flowUuid := message.FlowUuid

	go func() {
		select {
		case <-ctx.Done():
			slog.Debug(
				logging.FormatFlowPrefix(
					flowUuid,
					"closing by context",
				),
			)

			f.closing.Store(true)
		}
	}()

	go func(f *Feature) {
		for {
			select {
			case <-time.After(time.Second):
				if f.closing.Load() {
					break
				}

				if f.transport.IsConnected() {
					continue
				}

				slog.Debug(
					logging.FormatFlowPrefix(
						flowUuid,
						"closing by disconnected client",
					),
				)

				f.closing.Store(true)

				break
			}
		}
	}(f)

	err := f.transport.WriteResult(
		&dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Payload:  message.Payload,
			IsError:  false,
			Waitable: false,
		},
	)

	if err != nil {
		return &dto.Result{
			FlowUuid: flowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			IsError:  true,
			Payload: logging.FormatFlowPrefix(
				flowUuid,
				"can't send the handshake",
			),
		}
	}

	for {
		result := f.flows.PullResult(flowUuid)

		if result == nil {
			if f.closing.Load() {
				break
			}

			continue
		}

		slog.Debug(
			logging.FormatFlowTaskPrefix(
				flowUuid,
				result.TaskKey,
				"pull result",
			),
		)

		err := f.transport.WriteResult(result)

		if err != nil {
			slog.Error(
				logging.FormatFlowTaskPrefix(
					flowUuid,
					result.TaskKey,
					errs.Err(err).Error(),
				),
			)
		}
	}

	f.flows.Delete(flowUuid)

	return &dto.Result{
		FlowUuid: message.FlowUuid,
		Method:   message.Method,
		TaskKey:  message.TaskKey,
		Waitable: false,
	}
}
