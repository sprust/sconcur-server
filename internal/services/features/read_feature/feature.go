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
	"sync"
	"time"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
	flow      *flows.Flow
	transport *connection.Transport
}

func New(flow *flows.Flow, transport *connection.Transport) *Feature {
	return &Feature{
		flow:      flow,
		transport: transport,
	}
}

func (f *Feature) Handle(ctx context.Context, message *dto.Message) *dto.Result {
	flowUuid := message.FlowUuid

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

	slog.Debug(
		logging.FormatFlowPrefix(
			flowUuid,
			"handshake sent",
		),
	)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		results := f.flow.Listen()

		for {
			select {
			case <-ctx.Done():
				slog.Debug(
					logging.FormatFlowPrefix(
						flowUuid,
						"read: closing by context",
					),
				)

				return
			case <-ticker.C:
				slog.Debug(
					logging.FormatFlowPrefix(
						flowUuid,
						"check client connection",
					),
				)

				if f.transport.IsConnected() {
					continue
				}

				slog.Debug(
					logging.FormatFlowPrefix(
						flowUuid,
						"closing by disconnected client",
					),
				)

				return
			case result, ok := <-results:
				if !ok {
					slog.Debug(
						logging.FormatFlowTaskPrefix(
							flowUuid,
							result.TaskKey,
							"read: channel closed",
						),
					)
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
		}
	}()

	wg.Wait()

	f.flow.Stop()

	return &dto.Result{
		FlowUuid: message.FlowUuid,
		Method:   message.Method,
		TaskKey:  message.TaskKey,
		Waitable: false,
	}
}
