package crypto

import (
	"anubis-solver/polyfills"
	"crypto/sha256"
	"fmt"

	"rogchap.com/v8go"
)

type SubtleCrypto struct{}

var _ polyfills.Injector = (*SubtleCrypto)(nil)

func New() *SubtleCrypto {
	return &SubtleCrypto{}
}

func (s *SubtleCrypto) digest() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) != 2 {
			return nil
		}

		algorithm := args[0].String()
		if algorithm != "SHA-256" {
			return nil
		}

		arrayBuffer := args[1].ArrayBuffer()
		if !arrayBuffer.IsArrayBuffer() {
			return nil
		}

		resolver, err := v8go.NewPromiseResolver(info.Context())
		if err != nil {
			return nil
		}

		go func() {
			// Calculate SHA-256
			h := sha256.New()
			b := arrayBuffer.GetBytes()
			h.Write(b)
			result := h.Sum(nil)

			// Create ArrayBuffer with result
			buf := v8go.NewArrayBuffer(info.Context(), int64(len(result)))
			buf.PutBytes(result)

			// Resolve the promise
			resolver.Resolve(buf)
		}()

		return resolver.GetPromise().Value
	}
}

func (s *SubtleCrypto) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	// Create crypto object
	crypto := v8go.NewObjectTemplate(iso)

	// Create subtle object
	subtle := v8go.NewObjectTemplate(iso)

	// Add digest method to subtle
	digestFn := v8go.NewFunctionTemplate(iso, s.digest())
	subtle.Set("digest", digestFn)

	// Set subtle on crypto
	cryptoInstance, err := crypto.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("crypto object creation: %w", err)
	}

	subtleInstance, err := subtle.NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("subtle object creation: %w", err)
	}

	if err := cryptoInstance.Set("subtle", subtleInstance); err != nil {
		return fmt.Errorf("setting subtle: %w", err)
	}

	// Set crypto on global
	if err := global.Set("crypto", cryptoInstance); err != nil {
		return fmt.Errorf("setting crypto: %w", err)
	}

	return nil
}
