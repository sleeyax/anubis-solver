package base64

import (
	"anubis-solver/polyfills"
	stdBase64 "encoding/base64"
	"fmt"

	"rogchap.com/v8go"
)

type Base64 struct{}

var _ polyfills.Injector = (*Base64)(nil)

func New() *Base64 {
	return &Base64{}
}

func (b *Base64) atob() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			// TODO: v8go can't throw a error now, so we return an empty string
			return b.newStringValue(ctx, "")
		}

		encoded := args[0].String()

		byts, err := stdBase64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return b.newStringValue(ctx, "")
		}

		return b.newStringValue(ctx, string(byts))
	}
}

func (b *Base64) btoa() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		ctx := info.Context()

		if len(args) <= 0 {
			return b.newStringValue(ctx, "")
		}

		str := args[0].String()

		encoded := stdBase64.StdEncoding.EncodeToString([]byte(str))
		return b.newStringValue(ctx, encoded)
	}
}

func (b *Base64) newStringValue(ctx *v8go.Context, str string) *v8go.Value {
	iso := ctx.Isolate()
	val, _ := v8go.NewValue(iso, str)
	return val
}

func (b *Base64) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	for _, f := range []struct {
		Name string
		Func func() v8go.FunctionCallback
	}{
		{Name: "atob", Func: b.atob},
		{Name: "btoa", Func: b.btoa},
	} {
		template := v8go.NewFunctionTemplate(iso, f.Func())
		if err := global.Set(f.Name, template.GetFunction(ctx)); err != nil {
			return fmt.Errorf("base64 polyfill: %w", err)
		}
	}

	return nil
}
