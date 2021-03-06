package gornir

import (
	"context"
	"sync"
)

// TaskInstanceResult is the resut of running a task on a given host
type TaskInstanceResult interface{}

// Task is the interface that task plugins need to implement.
// the task is responsible to indicate its completion
// by calling sync.WaitGroup.Done()
type Task interface {
	Run(context.Context, Logger, *Host) (TaskInstanceResult, error)
}

// Runner is the interface of a struct that can implement a strategy
// to run tasks over hosts
type Runner interface {
	Run(context.Context, Logger, Task, map[string]*Host, chan *JobResult) error // Run executes the task over the hosts
	Close() error                                                               // Close closes and cleans all objects associated with the runner
	Wait() error                                                                // Wait blocks until all the hosts are done executing the task
}

// JobResult is the result of running a task over a host.
type JobResult struct {
	ctx  context.Context
	err  error
	host *Host
	data TaskInstanceResult
}

// NewJobResult instantiates a new JobResult
func NewJobResult(ctx context.Context, host *Host, data interface{}, err error) *JobResult {
	return &JobResult{
		ctx:  ctx,
		err:  err,
		host: host,
		data: data,
	}
}

// Context returns the context associated with the task
func (r *JobResult) Context() context.Context {
	return r.ctx
}

// Err returns the error the task set, otherwise nil
func (r *JobResult) Err() error {
	return r.err
}

// Host returns the host associated to the result
func (r *JobResult) Host() *Host {
	return r.host
}

// SetErr stores the error  and also propagates it to the associated Host
func (r *JobResult) SetErr(err error) {
	r.err = err
	r.host.SetErr(err)
}

// Data retrieves arbitrary data stored in the object
func (r *JobResult) Data() interface{} {
	return r.data
}

// SetData let's you store arbitrary data in the object
func (r *JobResult) SetData(data interface{}) {
	r.data = data
}

// TaskWrapper is a helper function that runs an instance of a task on a given host
func TaskWrapper(ctx context.Context, logger Logger, wg *sync.WaitGroup, taskFunc Task, host *Host, results chan *JobResult) {
	defer wg.Done()
	res, err := taskFunc.Run(ctx, logger, host)
	host.SetErr(err)
	results <- NewJobResult(ctx, host, res, err)
}
