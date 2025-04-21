package worker

import (
	"anubis-solver/polyfills"
	"rogchap.com/v8go"
)

type Worker struct{}

var _ polyfills.Injector = (*Worker)(nil)

func New() *Worker {
	return &Worker{}
}

func (b Worker) Inject(ctx *v8go.Context) error {
	_, err := ctx.RunScript(polyfill, "worker.js")

	return err
}
