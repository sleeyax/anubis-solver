package url

import (
	"anubis-solver/polyfills"
	"errors"

	"rogchap.com/v8go"
)

type URL struct{}

var _ polyfills.Injector = (*URL)(nil)

func New() *URL {
	return &URL{}
}

func (u *URL) Inject(ctx *v8go.Context) error {
	if ctx == nil {
		return errors.New("url polyfill: ctx is required")
	}

	_, err := ctx.RunScript(polyfill, "url-polyfill.js")

	return err
}
