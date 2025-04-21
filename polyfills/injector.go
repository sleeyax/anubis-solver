package polyfills

import "rogchap.com/v8go"

type Injector interface {
	Inject(ctx *v8go.Context) error
}
