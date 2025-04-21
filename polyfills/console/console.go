package console

import (
	"anubis-solver/polyfills"
	"fmt"
	"os"

	"rogchap.com/v8go"
)

type Console struct{}

var _ polyfills.Injector = (*Console)(nil)

func New() *Console {
	return &Console{}
}

func (c *Console) log() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if args := info.Args(); len(args) > 0 {
			inputs := make([]interface{}, len(args))
			for i, input := range args {
				inputs[i] = input
			}

			fmt.Fprintln(os.Stdout, inputs...)
		}

		return nil
	}
}

func (c *Console) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	console := v8go.NewObjectTemplate(iso)

	fn := v8go.NewFunctionTemplate(iso, c.log())

	methods := []string{
		"log",
		"debug",
		"error",
		"info",
		"warn",
		"assert",
	}
	for _, method := range methods {
		if err := console.Set(method, fn, v8go.ReadOnly); err != nil {
			return fmt.Errorf("console polyfill: %w", err)
		}
	}

	consoleInstance, err := console.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("console polyfill: %w", err)
	}

	global := ctx.Global()

	if err := global.Set("console", consoleInstance); err != nil {
		return fmt.Errorf("console polyfill: %w", err)
	}

	return nil
}
