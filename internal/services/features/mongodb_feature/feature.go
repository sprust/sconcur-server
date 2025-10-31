package mongodb_feature

import (
	"context"
	"encoding/json"
	"fmt"
	"sconcur/internal/services/contracts"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/features/mongodb_feature/connections"
	"sconcur/internal/services/features/mongodb_feature/helpers"
	"sconcur/internal/services/flows"
)

var _ contracts.MessageHandler = (*Feature)(nil)

type Feature struct {
}

func New() *Feature {
	return &Feature{}
}

func (f *Feature) Handle(flow *flows.Flow, message *dto.Message) *dto.Result {
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
				"mongodb: parse payload error: %s",
				err.Error(),
			),
		}
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	go func(ctxCancel context.CancelFunc) {
		select {
		case <-flow.StopListener():
			ctxCancel()
		}
	}(ctxCancel)

	connection := connections.NewConnections(ctx)

	collection, err := connection.Get(
		payload.Url,
		payload.Database,
		payload.Collection,
	)

	if err != nil {
		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: true,
			IsError:  true,
			Payload: fmt.Sprintf(
				"mongodb: %s",
				err.Error(),
			),
		}
	}

	if payload.Command == 1 {
		doc, err := helpers.UnmarshalDocument(payload.Data)

		if err != nil {
			return &dto.Result{
				FlowUuid: message.FlowUuid,
				Method:   message.Method,
				TaskKey:  message.TaskKey,
				Waitable: true,
				IsError:  true,
				Payload: fmt.Sprintf(
					"mongodb: parse payload data error: %s",
					err.Error(),
				),
			}
		}

		result, err := collection.InsertOne(ctx, doc)

		if err != nil {
			return &dto.Result{
				FlowUuid: message.FlowUuid,
				Method:   message.Method,
				TaskKey:  message.TaskKey,
				Waitable: true,
				IsError:  true,
				Payload: fmt.Sprintf(
					"mongodb: insert error: %s",
					err.Error(),
				),
			}
		}

		serializedResult, err := helpers.MarshalResult(result)

		if err != nil {
			return &dto.Result{
				FlowUuid: message.FlowUuid,
				Method:   message.Method,
				TaskKey:  message.TaskKey,
				Waitable: true,
				IsError:  true,
				Payload: fmt.Sprintf(
					"mongodb: marshal result error: %s",
					err.Error(),
				),
			}
		}

		return &dto.Result{
			FlowUuid: message.FlowUuid,
			Method:   message.Method,
			TaskKey:  message.TaskKey,
			Waitable: true,
			IsError:  false,
			Payload:  serializedResult,
		}
	}

	return &dto.Result{
		FlowUuid: message.FlowUuid,
		Method:   message.Method,
		TaskKey:  message.TaskKey,
		Waitable: true,
		IsError:  true,
		Payload:  "mongodb: unknow command",
	}
}
