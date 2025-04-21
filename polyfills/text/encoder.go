package text

import (
	"anubis-solver/polyfills"
	"fmt"
	"rogchap.com/v8go"
)

type Encoder struct{}

var _ polyfills.Injector = (*Encoder)(nil)

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) encode() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()
		iso := ctx.Isolate()

		if len(args) <= 0 {
			return nil
		}

		plaintext := args[0].String()
		uint8Array := []uint8(plaintext)

		v, _ := v8go.NewValue(iso, uint8Array)
		return v
	}
}

func (e *Encoder) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()

	constructorTemplate := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Constructor function
		instance, err := info.This().AsObject()
		if err != nil {
			return nil
		}

		// Set the encode method on the instance
		encodeFn := v8go.NewFunctionTemplate(iso, e.encode())
		fn := encodeFn.GetFunction(info.Context())
		instance.Set("encode", fn)

		return nil
	})

	constructor := constructorTemplate.GetFunction(ctx)

	global := ctx.Global()
	if err := global.Set("TextEncoder", constructor); err != nil {
		return fmt.Errorf("TextEncoder polyfill: %w", err)
	}

	return nil
}
