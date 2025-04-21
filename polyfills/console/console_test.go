package console

import (
	"testing"

	"rogchap.com/v8go"
)

func TestInject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)
	console := New()

	if err := console.Inject(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := ctx.RunScript("console.log(1111)", ""); err != nil {
		t.Error(err)
	}
}
