package read_feature

import (
	"context"
	"fmt"
	"log/slog"
	"sconcur/internal/services/connection"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/flows"
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
			slog.Debug("Read feature for flow [" + flowUuid + "] closing by context")

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

				slog.Debug("Read feature for flow [" + flowUuid + "] closing by disconnected client")

				f.closing.Store(true)

				break
			}
		}
	}(f)

	for {
		result := f.flows.PullResult(flowUuid)

		if result == nil {
			if f.closing.Load() {
				break
			}

			continue
		}

		slog.Debug(
			fmt.Sprintf(
				"Pull result for flow [%s] and task [%s]",
				flowUuid,
				result.TaskKey,
			),
		)

		err := f.transport.WriteResult(result)

		if err != nil {
			slog.Error(errs.Err(err).Error())
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
