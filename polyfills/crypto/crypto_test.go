package crypto

import (
	"rogchap.com/v8go"
	"testing"
)

func TestInject(t *testing.T) {
	iso := v8go.NewIsolate()
	ctx := v8go.NewContext(iso)

	if err := New().Inject(ctx); err != nil {
		t.Fatal(err)
	}

	v, err := ctx.RunScript(`var encoded = new Uint8Array([0,1,2]); var encoded = crypto.subtle.digest("SHA-256", encoded.buffer); encoded`, "cryto.js")

	if err != nil {
		t.Error(err)
	}

	if !v.IsPromise() {
		t.Errorf("Expected Promise return value")
	}

	promise, err := v.AsPromise()
	if err != nil {
		t.Error(err)
	}

	// Wait for the promise to resolve
	for {
		if promise.State() == v8go.Pending {
			continue
		}

		break
	}

	result := promise.Result()
	if !result.IsArrayBuffer() {
		t.Errorf("Expected promise to resolve to ArrayBuffer")
	}

	ab := result.ArrayBuffer()
	if ab.ByteLength() != 32 {
		t.Errorf("Got wrong ArrayBuffer length %d, expected 32", ab.ByteLength())
	}

}
