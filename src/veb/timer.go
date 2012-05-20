package veb

import (
	"time"
)

type Timer struct {
	start time.Time
	stop  time.Time
}

func (t *Timer) Start() {
	t.start = time.Now()
}

func (t *Timer) Stop() {
	t.stop = time.Now()
}

func (t *Timer) Duration() time.Duration {
	return t.stop.Sub(t.start)
}
