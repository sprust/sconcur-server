package sleep_feature

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/logging"
	"time"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
	Seconds int
}

func New() *Feature {
	return &Feature{}
}

func (s *Feature) Handle(ctx context.Context, message *dto.Message) *dto.Result {
	var payload Payload

	err := json.Unmarshal([]byte(message.Payload), &payload)

	if err != nil {
		slog.Warn(
			logging.FormatFlowTaskPrefix(
				message.FlowUuid,
				message.TaskKey,
				fmt.Sprintf(
					"sleep: parse error: %s",
					err.Error(),
				),
			),
		)

		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: true,
			IsError:  true,
			Payload: fmt.Sprintf(
				"parse error: %s",
				err.Error(),
			),
		}
	}

	select {
	case <-ctx.Done():
		slog.Warn(
			logging.FormatFlowTaskPrefix(
				message.FlowUuid,
				message.TaskKey,
				"sleep: closing by context",
			),
		)

		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: true,
			IsError:  true,
			Payload:  "closed by context",
		}
	case <-time.After(time.Duration(payload.Milliseconds) * time.Microsecond):
		slog.Debug(
			logging.FormatFlowTaskPrefix(
				message.FlowUuid,
				message.TaskKey,
				fmt.Sprintf(
					"sleep: woke up after [%dms]",
					payload.Milliseconds,
				),
			),
		)

		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: true,
			IsError:  false,
			Payload:  "",
		}
	}
}
