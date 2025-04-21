package window

import (
	"anubis-solver/polyfills"
	"rogchap.com/v8go"
)

type Window struct {
	location  string
	onReplace func(string)
}

var _ polyfills.Injector = (*Window)(nil)

func New(location string, onReplace func(string)) *Window {
	return &Window{location: location, onReplace: onReplace}
}

func (w Window) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	replaceFn := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()

		if len(args) > 0 {
			url := args[0].String()
			w.onReplace(url)
		}

		return nil
	})

	location := v8go.NewObjectTemplate(iso)
	if err := location.Set("href", w.location); err != nil {
		return err
	}
	if err := location.Set("replace", replaceFn); err != nil {
		return err
	}

	window := v8go.NewObjectTemplate(iso)
	if err := window.Set("location", location); err != nil {
		return err
	}
	if err := window.Set("isSecureContext", true); err != nil {
		return err
	}

	windowInstance, err := window.NewInstance(ctx)
	if err != nil {
		return err
	}

	if err := global.Set("window", windowInstance); err != nil {
		return err
	}

	return nil
}
