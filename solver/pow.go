package solver

import (
	"anubis-solver/polyfills"
	"anubis-solver/polyfills/base64"
	"anubis-solver/polyfills/blob"
	"anubis-solver/polyfills/console"
	"anubis-solver/polyfills/crypto"
	"anubis-solver/polyfills/document"
	"anubis-solver/polyfills/fetch"
	"anubis-solver/polyfills/listener"
	"anubis-solver/polyfills/navigator"
	"anubis-solver/polyfills/text"
	"anubis-solver/polyfills/timers"
	"anubis-solver/polyfills/url"
	"anubis-solver/polyfills/window"
	"anubis-solver/polyfills/worker"
	"fmt"
	v8 "rogchap.com/v8go"
	"strings"
)

type Window struct {
	v8.Object
}

func SolveChallenge(uri, source, version, challenge string) (string, error) {
	source = strings.ReplaceAll(source, "setTimeout", "finish") // setTimeout is kinda broken rn so we just replace it with a custom function for now

	resultChan := make(chan string, 1)

	iso := v8.NewIsolate()
	ctx := v8.NewContext(iso)

	// Set up event listeners
	in := make(chan *v8.Object)
	out := make(chan *v8.Value)
	go func() {
		for {
			select {
			case v := <-out:
				fmt.Println("out:", v)
			}
		}
	}()
	go func() {
		for {
			select {
			case v := <-in:
				fmt.Println("in:", v)
			}
		}
	}()
	listenerInjector := listener.New()
	if err := listenerInjector.AddListeners("message", in, out); err != nil {
		return "", fmt.Errorf("failed to add listeners: %w", err)
	}

	// Inject polyfills
	injectors := []polyfills.Injector{
		url.New(),
		base64.New(),
		crypto.New(),
		text.NewEncoder(),
		fetch.New(),
		console.New(),
		text.NewDecoder(),
		timers.New(),
		blob.New(), // depends on text.NewDecoder()
		listenerInjector,
		worker.New(), // depends on event listeners
		window.New(uri, func(url string) {
			resultChan <- url
		}),
		document.New(fmt.Sprintf("\"%s\"", version), challenge),
		navigator.New(navigator.Options{
			Threads: 1,
		}),
	}
	for _, injector := range injectors {
		if err := injector.Inject(ctx); err != nil {
			return "", err
		}
	}

	finishFn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) > 0 && args[0].IsFunction() {
			fn, _ := args[0].AsFunction()
			result, _ := fn.Call(v8.Undefined(iso))
			return result
		}
		return nil
	})

	if err := ctx.Global().Set("finish", finishFn.GetFunction(ctx)); err != nil {
		return "", fmt.Errorf("failed to set finish function: %w", err)
	}

	// Run the source script
	_, err := ctx.RunScript(source, "main.mjs")
	if err != nil {
		return "", fmt.Errorf("failed to run script: %w", err)
	}

	result := <-resultChan
	return result, nil

	/*	promise, err := val.AsPromise()
		if err != nil {
			return err
		}

		for {
			if promise.State() == v8.Pending {
				continue
			}

			break
		}

		val = promise.Result()

		v, _ := val.MarshalJSON()

		fmt.Println(val)
		fmt.Println(string(v))*/
}
