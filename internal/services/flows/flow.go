package flows

import (
	"context"
	"fmt"
	"log/slog"
	"sconcur/internal/services/connection"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/logging"
	"sconcur/pkg/foundation/errs"
	"sync"
	"time"
)

// TODO: FIFO maps

type Flow struct {
	mutex sync.Mutex

	flowUuid string

	transport *connection.Transport

	active map[string]*dto.Task

	totalCount     int
	activeCount    int
	completedCount int

	resultsChannel chan *dto.Result

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewFlow(flowUuid string, transport *connection.Transport) *Flow {
	ctx, ctxCancel := context.WithCancel(context.Background())

	return &Flow{
		flowUuid: flowUuid,

		transport: transport,

		active: make(map[string]*dto.Task),

		ctx:       ctx,
		ctxCancel: ctxCancel,

		resultsChannel: make(chan *dto.Result),
	}
}

func (f *Flow) Run() error {
	err := f.transport.WriteResult(
		&dto.Result{
			FlowUuid: f.flowUuid,
			Method:   1,
			IsError:  false,
		},
	)

	if err != nil {
		return errs.Err(err)
	}

	slog.Debug(
		logging.FormatFlowPrefix(
			f.flowUuid,
			"handshake sent",
		),
	)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if f.transport.IsConnected() {
					continue
				}

				slog.Debug(
					logging.FormatFlowPrefix(
						f.flowUuid,
						"closing by disconnected client",
					),
				)

				return
			}
		}
	}()

	wg.Wait()

	f.ctxCancel()

	return nil
}

func (f *Flow) AddMessage(message *dto.Message) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.active[message.TaskKey] = &dto.Task{
		Message: message,
		Result:  nil,
	}

	f.totalCount++
	f.activeCount++
}

func (f *Flow) AddResult(result *dto.Result) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	task, exists := f.active[result.TaskKey]

	if !exists {
		return errs.Err(
			fmt.Errorf(
				"task with key [%s] not found in flow [%s]",
				result.TaskKey,
				result.FlowUuid,
			),
		)
	}

	task.Result = result

	delete(f.active, result.TaskKey)
	f.activeCount--

	f.completedCount++

	err := f.transport.WriteResult(result)

	if err != nil {
		return errs.Err(err)
	}

	slog.Debug(
		logging.FormatFlowTaskPrefix(
			result.FlowUuid,
			result.TaskKey,
			"pushed result to channel",
		),
	)

	return nil
}

func (f *Flow) GetTotalCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.totalCount
}

func (f *Flow) StopListener() <-chan struct{} {
	return f.ctx.Done()
}
