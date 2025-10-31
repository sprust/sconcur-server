package sleep_feature

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/flows"
	"sconcur/internal/services/logging"
	"time"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
}

func New() *Feature {
	return &Feature{}
}

func (s *Feature) Handle(flow *flows.Flow, message *dto.Message) *dto.Result {
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
	case <-flow.StopListener():
		slog.Debug(
			logging.FormatFlowTaskPrefix(
				message.FlowUuid,
				message.TaskKey,
				"sleep: closing by flow stop",
			),
		)

		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: false,
			IsError:  true,
			Payload:  "closed by flow stop",
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
