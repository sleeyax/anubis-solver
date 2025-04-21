package fetch

import (
	"anubis-solver/polyfills"
	"fmt"
	"rogchap.com/v8go"
)

var _ polyfills.Injector = (*Fetcher)(nil)

func (f *Fetcher) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	template := v8go.NewFunctionTemplate(iso, f.fetch())

	if err := global.Set("fetch", template.GetFunction(ctx)); err != nil {
		return fmt.Errorf("fetch polyfill: %w", err)
	}

	return nil
}
