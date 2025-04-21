package timers

import (
	"anubis-solver/polyfills/console"
	"testing"
	"time"

	"rogchap.com/v8go"
)

func Test_SetTimeout(t *testing.T) {
	ctx, err := newV8ContextWithTimers()
	if err != nil {
		t.Fatal(err)
	}

	if err := console.New().Inject(ctx); err != nil {
		t.Fatal(err)
	}

	val, err := ctx.RunScript(`
	console.log(new Date().toUTCString());

	setTimeout(function() {
		console.log("Hello v8go.");
		console.log(new Date().toUTCString());
	}, 2000)`, "set_timeout.js")
	if err != nil {
		t.Fatal(err)
	}

	if !val.IsInt32() {
		t.Fatalf("except 1 but got %v", val)
	}

	if id := val.Int32(); id != 1 {
		t.Fatalf("except 1 but got %d", id)
	}

	time.Sleep(time.Second * 6)
}

func newV8ContextWithTimers() (*v8go.Context, error) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	if err := New().Inject(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}
