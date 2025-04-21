package navigator

import (
	"anubis-solver/polyfills"
	"rogchap.com/v8go"
)

type Navigator struct {
	options Options
}

type Options struct {
	Threads int32
}

var _ polyfills.Injector = (*Navigator)(nil)

func New(options Options) *Navigator {
	return &Navigator{
		options: options,
	}
}

func (n *Navigator) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	// Set 'navigator.hardwareConcurrency'
	navigator := v8go.NewObjectTemplate(iso)
	if err := navigator.Set("hardwareConcurrency", n.options.Threads); err != nil {
		return err
	}

	navigatorInstance, err := navigator.NewInstance(ctx)
	if err != nil {
		return err
	}

	if err := global.Set("navigator", navigatorInstance); err != nil {
		return err
	}

	return nil
}
