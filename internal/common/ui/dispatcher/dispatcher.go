package dispatcher

import (
	"context"
	"sync"

	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/common/ui/uimodels"
	"github.com/pkg/errors"
)

type Dispatcher struct {
	mutex     *sync.Mutex
	runningOp uimodels.Operation
	publisher MessagePublisher
}

func NewDispatcher(publisher MessagePublisher) *Dispatcher {
	return &Dispatcher{
		mutex:     new(sync.Mutex),
		publisher: publisher,
	}
}

func (d *Dispatcher) Start(ctx context.Context, operation uimodels.Operation) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.runningOp != nil {
		d.publisher.Send(events.Error(errors.New("operation already running")))
	}

	d.runningOp = operation
	go func() {
		subCtx := uimodels.WithContext(ctx, DispatcherContext{d.publisher})

		err := operation.Execute(subCtx)
		if err != nil {
			d.publisher.Send(events.Error(err))
		}

		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.runningOp = nil
	}()
}
