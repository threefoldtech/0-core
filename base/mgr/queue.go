package mgr

import (
	"container/list"
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
}

func (q *Queue) Init() {
	q.o.Do(func() {
		q.queues = make(map[string]*list.List)
		q.ch = make(chan *jobImb)
	})
}

func (q *Queue) Channel() <-chan *jobImb {
	return q.ch
}

func (q *Queue) Push(job *jobImb) {
	q.lock.Lock()
	defer q.lock.Unlock()

	name := job.Command().Queue
	if name == "" {
		q.ch <- job
		return
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
}

func (q *Queue) Notify(job pm.Job) {
	q.lock.Lock()
	defer q.lock.Unlock()
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
