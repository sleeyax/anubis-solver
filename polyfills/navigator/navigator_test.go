package navigator

import (
	"rogchap.com/v8go"
	"testing"
)

func TestNavigator_Inject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)
	var numThreads int32 = 24

	if err := New(Options{Threads: numThreads}).Inject(ctx); err != nil {
		t.Fatalf("failed to inject navigator: %v", err)
	}

	res, err := ctx.RunScript("navigator.hardwareConcurrency", "navigator.js")
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	if !res.IsInt32() {
		t.Fatalf("expected int32, got %T", res)
	}

	if res.Int32() != numThreads {
		t.Fatalf("expected 24, got %d", res.Int32())
	}
}
