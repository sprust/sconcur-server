package flows

import (
	"errors"
	"log/slog"
	"sconcur/internal/services/dto"
	"sconcur/internal/services/logging"
	"sconcur/pkg/foundation/errs"
	"sync"
)

type Flows struct {
	mutex sync.RWMutex
	flows map[string]*Flow
}

func NewFlows() *Flows {
	return &Flows{
		flows: make(map[string]*Flow),
	}
}

func (f *Flows) Add(flowUuid string, flow *Flow) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.flows[flowUuid] = flow

	go func(flows *Flows, flowUuid string, flow *Flow) {
		<-flow.StopListener()

		slog.Debug(
			logging.FormatFlowPrefix(
				flowUuid,
				"flows closing flow by flow stop",
			),
		)

		flows.delete(flowUuid)
	}(f, flowUuid, flow)
}

func (f *Flows) Get(flowUuid string) (*Flow, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	flow, exists := f.flows[flowUuid]

	if !exists {
		return nil, errors.New(
			logging.FormatFlowPrefix(
				flowUuid,
				"flow not found at Get",
			),
		)
	}

	return flow, nil
}

func (f *Flows) AddMessage(message *dto.Message) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	flow, exists := f.flows[message.FlowUuid]

	if !exists {
		return errs.Err(
			errors.New(
				logging.FormatFlowTaskPrefix(
					message.FlowUuid,
					message.TaskKey,
					"flow not found at AddMessage",
				),
			),
		)
	}

	flow.AddMessage(message)

	return nil
}

func (f *Flows) AddResult(result *dto.Result) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	flow, exists := f.flows[result.FlowUuid]

	if !exists {
		return errs.Err(
			errors.New(
				logging.FormatFlowTaskPrefix(
					result.FlowUuid,
					result.TaskKey,
					"flow not found at AddResult",
				),
			),
		)
	}

	// TODO: handle error
	_ = flow.AddResult(result)

	return nil
}

func (f *Flows) GetCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.flows)
}

func (f *Flows) delete(flowUuid string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	delete(f.flows, flowUuid)
}
