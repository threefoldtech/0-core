package mgr

import (
	"sync"
	"time"
)

//BackOff a back off strategy
type BackOff struct {
	Delay    time.Duration
	Multiply float64
	Max      time.Duration

	lastDelay time.Duration
	lastCall  time.Time

	o sync.Once
}

func (b *BackOff) defaults() {
	b.lastCall = time.Now()
	if b.Delay == time.Duration(0) {
		b.Delay = 30 * time.Second
	}

	if b.Multiply == 0 {
		b.Multiply = 2
	}

	if b.Max == time.Duration(0) {
		b.Max = 10 * time.Minute
	}
}

//Duration get the next backoff duration according to policy
func (b *BackOff) Duration() (duration time.Duration) {
	b.o.Do(b.defaults)
	now := time.Now()

	defer func() {
		b.lastCall = now
		duration = (duration / time.Second) * time.Second
		if duration >= b.Max {
			duration = b.Max
		}

		b.lastDelay = duration
	}()

	since := now.Sub(b.lastCall)
	//since can be very small as small as 0

	if since < b.lastDelay*2 {
		//call to duration was too fast
		//apply policy
		duration = time.Duration(float64(b.lastDelay) * b.Multiply)
		return
	}
	//reset
	duration = b.Delay
	return
}
