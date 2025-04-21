package text

import (
	"testing"

	"rogchap.com/v8go"
)

func TestDecoder_Inject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)
	decoder := NewDecoder()

	if err := decoder.Inject(ctx); err != nil {
		t.Fatal(err)
	}

	val, err := ctx.RunScript("new TextDecoder().decode(new Uint8Array([97,98,99]))", "decode.js")
	if err != nil {
		t.Fatal(err)
	}

	if !val.IsString() {
		t.Fatalf("Expected string, got %s", val)
	}

	if val.String() != "abc" {
		t.Fatalf("Expected 'abc', got %s", val.String())
	}
}
