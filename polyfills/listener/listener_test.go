package listener_test

import (
	"anubis-solver/polyfills/listener"
	"github.com/stretchr/testify/assert"
	"testing"

	v8 "rogchap.com/v8go"
)

func BenchmarkEventListenerCall(b *testing.B) {
	iso := v8.NewIsolate()
	ctx := v8.NewContext(iso)

	in := make(chan *v8.Object)
	out := make(chan *v8.Value)
	l := listener.New()
	_ = l.AddListeners("auth", in, out)

	if err := l.Inject(ctx); err != nil {
		b.Fatal(err)
	}

	_, err := ctx.RunScript("addListener('auth', event => { return event.sourceIP === '127.0.0.1' })", "listener.js")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		obj, err := newContextObject(ctx)
		assert.NoError(b, err)
		in <- obj

		v := <-out

		assert.NotNil(b, v)
		assert.True(b, v.IsBoolean())
	}
}

func newContextObject(ctx *v8.Context) (*v8.Object, error) {
	iso := ctx.Isolate()
	obj := v8.NewObjectTemplate(iso)

	resObj, err := obj.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range []struct {
		Key string
		Val interface{}
	}{
		{Key: "sourceIP", Val: "127.0.0.1"},
	} {
		if err := resObj.Set(v.Key, v.Val); err != nil {
			return nil, err
		}
	}

	return resObj, nil
}
