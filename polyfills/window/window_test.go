package window

import (
	"rogchap.com/v8go"
	"testing"
)

func TestWindow_Inject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)
	location := "https://example.com/script.js"

	if err := New(location).Inject(ctx); err != nil {
		t.Fatal(err)
	}

	v, err := ctx.RunScript(`window.location.href`, "cryto.js")

	if err != nil {
		t.Fatal(err)
	}

	if !v.IsString() {
		t.Errorf("Expected string return value")
	}

	if v.String() != location {
		t.Errorf("Expected %s, got %s", location, v.String())
	}
}
