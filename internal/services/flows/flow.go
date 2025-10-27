package flows

import (
	"fmt"
	"sconcur/internal/services/dto"
	"sconcur/pkg/foundation/errs"
	"sync"
)

// TODO: FIFO maps

type Flow struct {
	mutex sync.Mutex

	active    map[string]*dto.Task
	completed map[string]*dto.Task

	totalCount     int
	activeCount    int
	completedCount int
}

func NewFlow() *Flow {
	return &Flow{
		active:    make(map[string]*dto.Task),
		completed: make(map[string]*dto.Task),
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

	f.completed[result.TaskKey] = task
	f.completedCount++

	return nil
}

func (f *Flow) PullResult() *dto.Task {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for key, task := range f.completed {
		delete(f.completed, key)
		f.completedCount--
		f.totalCount--

		return task
	}

	return nil
}

func (f *Flow) GetTotalCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.totalCount
}
