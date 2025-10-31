package flows

import (
	"errors"
	"sconcur/internal/services/logging"
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

func (f *Flows) GetCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.flows)
}

func (f *Flows) Delete(flowUuid string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	delete(f.flows, flowUuid)
}
