package flows

import (
	"context"
	"fmt"
	"log/slog"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/logging"
	"sconcur/pkg/foundation/errs"
	"sync"
)

// TODO: FIFO maps

type Flow struct {
	mutex sync.Mutex

	active map[string]*dto.Task

	totalCount     int
	activeCount    int
	completedCount int

	resultsChannel chan *dto.Result

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewFlow() *Flow {
	ctx, ctxCancel := context.WithCancel(context.Background())

	return &Flow{
		active: make(map[string]*dto.Task),

		ctx:       ctx,
		ctxCancel: ctxCancel,

		resultsChannel: make(chan *dto.Result),
	}
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

	f.resultsChannel <- result

	slog.Debug(
		logging.FormatFlowTaskPrefix(
			result.FlowUuid,
			result.TaskKey,
			"pushed result to channel",
		),
	)

	return nil
}

func (f *Flow) Listen() <-chan *dto.Result {
	return f.resultsChannel
}

func (f *Flow) GetTotalCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.totalCount
}

func (f *Flow) Stop() {
	f.ctxCancel()
}

func (f *Flow) StopListener() <-chan struct{} {
	return f.ctx.Done()
}
