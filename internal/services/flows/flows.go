package flows

import (
	"sconcur/internal/services/dto"
	"sync"
)

var once sync.Once
var instance *Flows

type Flows struct {
	mutex sync.Mutex
	flows map[string]*Flow
}

func GetFlows() *Flows {
	once.Do(func() {
		instance = &Flows{
			flows: make(map[string]*Flow),
		}
	})

	return instance
}

func (f *Flows) AddMessage(message *dto.Message) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	flow, exists := f.flows[message.FlowUuid]

	if !exists {
		flow = NewFlow()

		f.flows[message.FlowUuid] = flow
	}

	flow.AddMessage(message)
}

func (f *Flows) AddResult(result *dto.Result) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	flow, exists := f.flows[result.FlowUuid]

	if !exists {
		flow = NewFlow()

		f.flows[result.FlowUuid] = flow
	}

	// TODO: handle error
	_ = flow.AddResult(result)
}

func (f *Flows) PullResult(flowUuid string) *dto.Result {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	flow, exists := f.flows[flowUuid]

	if !exists {
		return nil
	}

	task := flow.PullResult()

	if task == nil {
		return nil
	}

	if flow.GetTotalCount() == 0 {
		delete(f.flows, flowUuid)
	}

	return task.Result
}

func (f *Flows) Delete(flowUuid string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	delete(f.flows, flowUuid)
}

func (f *Flows) GetCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return len(f.flows)
}
