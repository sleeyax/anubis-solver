package document

import (
	"rogchap.com/v8go"
	"testing"
)

func TestWindow_Inject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	if err := New(`"v1.16.0-24-g75b97eb"`).Inject(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := ctx.RunScript(`const container = document.createElement("div");  container.style.marginTop = "1rem";  container.style.display = "flex"; container.title = "hello";`, "document.js")

	if err != nil {
		t.Fatal(err)
	}

	_, err = ctx.RunScript(`const containerById = document.getElementById("container");`, "document.js")

	if err != nil {
		t.Fatal(err)
	}
}
