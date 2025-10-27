package sleep_feature

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
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
			fmt.Sprintf(
				"Sleep feature for flow [%s] and task [%s] closing by context",
				message.FlowUuid,
				message.TaskKey,
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
			fmt.Sprintf(
				"Sleep feature for flow [%s] and task [%s] woke up after [%dms]",
				message.FlowUuid,
				message.TaskKey,
				payload.Milliseconds,
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
