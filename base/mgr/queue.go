package mgr

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/threefoldtech/0-core/base/pm"
)

/**
Queue is used for sequential cmds exectuions
*/
type Queue struct {
	queues map[string]*list.List
	ch     chan *jobImb
	lock   sync.Mutex
	o      sync.Once
	closed bool
}

//Init initializes the queue
func (q *Queue) Init() {
	q.o.Do(func() {
		q.queues = make(map[string]*list.List)
		q.ch = make(chan *jobImb)
	})
}

//Close the queue. Queue can't be used after close
func (q *Queue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true
	close(q.ch)
}

//Channel return job channel
func (q *Queue) Channel() <-chan *jobImb {
	return q.ch
}

//Push a job on queue
func (q *Queue) Push(job *jobImb) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return fmt.Errorf("closed queue")
	}

	name := job.Command().Queue
	if name == "" {
		q.ch <- job
		return nil
	}

	queue, ok := q.queues[name]
	if !ok {
		queue = list.New()
		q.queues[name] = queue
	}

	queue.PushBack(job)
	if queue.Len() == 1 {
		//first job in the queue
		q.ch <- job
	}

	return nil
}

//Notify tell queue that a job execution has completed
func (q *Queue) Notify(job pm.Job) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}

	name := job.Command().Queue
	queue, ok := q.queues[name]
	if !ok {
		return
	}
	queue.Remove(queue.Front())
	if queue.Len() == 0 {
		delete(q.queues, name)
		return
	}

	next := queue.Front().Value.(*jobImb)
	q.ch <- next
}
