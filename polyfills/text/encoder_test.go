package text

import (
	"testing"

	"rogchap.com/v8go"
)

func TestEncoder_Inject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)
	encoder := NewEncoder()

	if err := encoder.Inject(ctx); err != nil {
		t.Fatal(err)
	}

	val, err := ctx.RunScript("new TextEncoder().encode('abc')", "encode.js")
	if err != nil {
		t.Fatal(err)
	}

	if s := val.String(); s != "97,98,99" {
		t.Fatalf("Expected Uint8Array as string, got %s", s)
	}
}
