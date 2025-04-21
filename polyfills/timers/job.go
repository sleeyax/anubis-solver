package timers

import (
	"time"
)

type FunctionCallback func()

type ClearCallback func(id int32)

type Job struct {
	ID            int32
	Done          bool
	Cleared       bool
	Interval      bool
	Delay         int32
	ClearCallback ClearCallback
	FunctionCB    FunctionCallback
}

func (j *Job) Clear() {
	if !j.Cleared {
		j.Cleared = true

		if j.ClearCallback != nil {
			j.ClearCallback(j.ID)
		}
	}

	j.Done = true
}

func (j *Job) Start() {
	go func() {
		defer j.Clear()

		ticker := time.NewTicker(time.Duration(j.Delay) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if j.Done {
				break
			}

			if j.FunctionCB != nil {
				j.FunctionCB()
			}

			if !j.Interval {
				j.Done = true
				break
			}
		}
	}()
}
