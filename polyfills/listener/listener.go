package listener

import (
	"anubis-solver/polyfills"
	"fmt"
	v8 "rogchap.com/v8go"
)

type Listener struct {
	in  map[string]chan *v8.Object
	out map[string]chan *v8.Value
}

var _ polyfills.Injector = (*Listener)(nil)

func New() *Listener {
	l := new(Listener)
	l.in = make(map[string]chan *v8.Object)
	l.out = make(map[string]chan *v8.Value)
	return l
}

func (l *Listener) AddInputListener(name string, chn chan *v8.Object) error {
	if _, ok := l.in[name]; ok {
		return fmt.Errorf("input listener '%s' already exists", name)
	}

	l.in[name] = chn

	return nil
}

func (l *Listener) AddOutputListener(name string, chn chan *v8.Value) error {
	if _, ok := l.out[name]; ok {
		return fmt.Errorf("output listener '%s' already exists", name)
	}

	l.out[name] = chn

	return nil
}

func (l *Listener) RemoveInputListener(name string) {
	if _, ok := l.in[name]; ok {
		delete(l.in, name)
	}
}

func (l *Listener) RemoveOutputListener(name string) {
	if _, ok := l.out[name]; ok {
		delete(l.out, name)
	}
}

func (l *Listener) AddListeners(name string, in chan *v8.Object, out chan *v8.Value) error {
	if err := l.AddInputListener(name, in); err != nil {
		return err
	}
	if err := l.AddOutputListener(name, out); err != nil {
		return err
	}
	return nil
}

func (l *Listener) RemoveListeners(name string) {
	l.RemoveInputListener(name)
	l.RemoveOutputListener(name)
}

func (l *Listener) Inject(ctx *v8.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	template := v8.NewFunctionTemplate(iso, l.addListener())

	if err := global.Set("addListener", template.GetFunction(ctx)); err != nil {
		return fmt.Errorf("listener polyfill: %w", err)
	}
	if err := global.Set("addEventListener", template.GetFunction(ctx)); err != nil {
		return fmt.Errorf("listener polyfill: %w", err)
	}

	return nil
}

func (l *Listener) addListener() v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		ctx := info.Context()
		args := info.Args()

		if len(args) <= 1 {
			err := fmt.Errorf("addListener: expected 2 arguments, got %d", len(args))

			return l.newErrorValue(ctx, err)
		}

		fn, err := args[1].AsFunction()
		if err != nil {
			err := fmt.Errorf("%w", err)

			return l.newErrorValue(ctx, err)
		}

		chn, ok := l.in[args[0].String()]
		if !ok {
			err := fmt.Errorf("addListener: event '%s' not found", args[0].String())

			return l.newErrorValue(ctx, err)
		}

		go func(chn chan *v8.Object, fn *v8.Function) {
			for e := range chn {
				v, err := fn.Call(ctx.Global(), e)
				if err != nil {
					fmt.Printf("addListener: %v", err)
				}

				l.out[args[0].String()] <- v
			}
		}(chn, fn)

		return v8.Undefined(ctx.Isolate())
	}
}

func (l *Listener) newErrorValue(ctx *v8.Context, err error) *v8.Value {
	iso := ctx.Isolate()
	e, _ := v8.NewValue(iso, fmt.Sprintf("addListener: %v", err))
	return e
}
