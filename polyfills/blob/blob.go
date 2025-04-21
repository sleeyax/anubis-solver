package blob

import (
	"anubis-solver/polyfills"
	"rogchap.com/v8go"
)

type Blob struct{}

var _ polyfills.Injector = (*Blob)(nil)

func New() *Blob {
	return &Blob{}
}

func (b Blob) Inject(ctx *v8go.Context) error {
	_, err := ctx.RunScript(polyfill, "blob.js")

	return err
}
