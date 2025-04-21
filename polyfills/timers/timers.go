package timers

import (
	"anubis-solver/polyfills"
	"errors"
	"fmt"

	"rogchap.com/v8go"
)

type Timers struct {
	jobs      map[int32]*Job
	nextJobID int32
}

var _ polyfills.Injector = (*Timers)(nil)

func New() *Timers {
	return &Timers{
		jobs:      make(map[int32]*Job),
		nextJobID: 1,
	}
}

func (t *Timers) GetSetTimeoutFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()

		id, err := t.startNewTimer(info.This(), info.Args(), false)
		if err != nil {
			return t.newInt32Value(ctx, 0)
		}

		return t.newInt32Value(ctx, id)
	}
}

func (t *Timers) GetSetIntervalFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()

		id, err := t.startNewTimer(info.This(), info.Args(), true)
		if err != nil {
			return t.newInt32Value(ctx, 0)
		}

		return t.newInt32Value(ctx, id)
	}
}

func (t *Timers) GetClearTimeoutFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) > 0 && args[0].IsInt32() {
			t.clear(args[0].Int32(), false)
		}

		return nil
	}
}

func (t *Timers) GetClearIntervalFunctionCallback() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) > 0 && args[0].IsInt32() {
			t.clear(args[0].Int32(), true)
		}

		return nil
	}
}

func (t *Timers) clear(id int32, interval bool) {
	if id < 1 {
		return
	}

	if item, ok := t.jobs[id]; ok && item.Interval == interval {
		item.Clear()
	}
}

func (t *Timers) startNewTimer(this v8go.Valuer, args []*v8go.Value, interval bool) (int32, error) {
	if len(args) <= 0 {
		return 0, errors.New("1 argument required, but only 0 present")
	}

	fn, err := args[0].AsFunction()
	if err != nil {
		return 0, err
	}

	var delay int32
	if len(args) > 1 && args[1].IsInt32() {
		delay = args[1].Int32()
	}
	if delay < 10 {
		delay = 10
	}

	var restArgs []v8go.Valuer
	if len(args) > 2 {
		restArgs = make([]v8go.Valuer, 0)
		for _, arg := range args[2:] {
			restArgs = append(restArgs, arg)
		}
	}

	item := &Job{
		ID:       t.nextJobID,
		Done:     false,
		Cleared:  false,
		Delay:    delay,
		Interval: interval,
		FunctionCB: func() {
			_, _ = fn.Call(this, restArgs...)
		},
		ClearCallback: func(id int32) {
			delete(t.jobs, id)
		},
	}

	t.nextJobID++
	t.jobs[item.ID] = item

	item.Start()

	return item.ID, nil
}

func (t *Timers) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	for _, f := range []struct {
		Name string
		Func func() v8go.FunctionCallback
	}{
		{Name: "setTimeout", Func: t.GetSetTimeoutFunctionCallback},
		{Name: "setInterval", Func: t.GetSetIntervalFunctionCallback},
		{Name: "clearTimeout", Func: t.GetClearTimeoutFunctionCallback},
		{Name: "clearInterval", Func: t.GetClearIntervalFunctionCallback},
	} {
		template := v8go.NewFunctionTemplate(iso, f.Func())

		if err := global.Set(f.Name, template.GetFunction(ctx)); err != nil {
			return fmt.Errorf("timers polyfill: %w", err)
		}
	}

	return nil
}

func (t *Timers) newInt32Value(ctx *v8go.Context, i int32) *v8go.Value {
	iso := ctx.Isolate()
	v, _ := v8go.NewValue(iso, i)
	return v
}
