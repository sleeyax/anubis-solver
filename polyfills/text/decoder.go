package text

import (
	"anubis-solver/polyfills"
	"fmt"
	"rogchap.com/v8go"
)

type Decoder struct{}

var _ polyfills.Injector = (*Decoder)(nil)

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) decode() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()
		iso := ctx.Isolate()

		if len(args) <= 0 {
			return nil
		}

		// Extract Uint8Array from the first argument
		uint8Array := args[0].Uint8Array()

		// Convert bytes to string
		str := string(uint8Array)
		v, _ := v8go.NewValue(iso, str)
		return v
	}
}

func (d *Decoder) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()

	constructorTemplate := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		instance, err := info.This().AsObject()
		if err != nil {
			return nil
		}

		// Set the decode method on the instance
		decodeFn := v8go.NewFunctionTemplate(iso, d.decode())
		fn := decodeFn.GetFunction(info.Context())
		instance.Set("decode", fn)

		return nil
	})

	constructor := constructorTemplate.GetFunction(ctx)

	global := ctx.Global()
	if err := global.Set("TextDecoder", constructor); err != nil {
		return fmt.Errorf("TextDecoder polyfill: %w", err)
	}

	return nil
}
