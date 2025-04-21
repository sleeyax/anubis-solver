package document

import (
	"anubis-solver/polyfills"
	"rogchap.com/v8go"
)

type Document struct {
	rawAnubisVersion   string
	rawAnubisChallenge string
}

var _ polyfills.Injector = (*Document)(nil)

func New(rawAnubisVersion string, rawAnubisChallenge string) *Document {
	return &Document{rawAnubisVersion: rawAnubisVersion, rawAnubisChallenge: rawAnubisChallenge}
}

func (d *Document) createDummyObject(ctx *v8go.Context) *v8go.Value {
	obj := v8go.NewObjectTemplate(ctx.Isolate())

	dummyObjectInstance, err := obj.NewInstance(ctx)
	if err != nil {
		return nil
	}

	styleObject := v8go.NewObjectTemplate(ctx.Isolate())
	styleObjectInstance, _ := styleObject.NewInstance(ctx)
	if err := dummyObjectInstance.Set("style", styleObjectInstance); err != nil {
		return nil
	}

	return dummyObjectInstance.Value
}

func (d *Document) createElement() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		return d.createDummyObject(info.Context())
	}
}

func (d *Document) getElementById() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		elementName := args[0].String()

		dummyObject := v8go.NewObjectTemplate(info.Context().Isolate())

		if err := dummyObject.Set("appendChild", v8go.NewFunctionTemplate(info.Context().Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			return nil
		})); err != nil {
			return nil
		}

		switch elementName {
		case "anubis_version":
			if err := dummyObject.Set("textContent", d.rawAnubisVersion); err != nil {
				return nil
			}
		case "progress":
			// Create progress element with style
			progressInstance, err := dummyObject.NewInstance(info.Context())
			if err != nil {
				return nil
			}

			// Create style for progress element
			progressStyle := v8go.NewObjectTemplate(info.Context().Isolate())
			progressStyle.Set("display", "")
			progressStyleInstance, _ := progressStyle.NewInstance(info.Context())
			progressInstance.Set("style", progressStyleInstance)

			// Create style for child element
			childStyle := v8go.NewObjectTemplate(info.Context().Isolate())
			childStyle.Set("width", "")
			childStyleInstance, _ := childStyle.NewInstance(info.Context())

			// Create child element with style
			childElement := v8go.NewObjectTemplate(info.Context().Isolate())
			childInstance, _ := childElement.NewInstance(info.Context())
			childInstance.Set("style", childStyleInstance)

			// Set child element
			progressInstance.Set("firstElementChild", childInstance)
			progressInstance.Set("aria-valuenow", 0)

			return progressInstance.Value
		case "anubis_challenge":
			if err := dummyObject.Set("textContent", d.rawAnubisChallenge); err != nil {
				return nil
			}
		}

		dummyObjectInstance, err := dummyObject.NewInstance(info.Context())
		if err != nil {
			return nil
		}

		return dummyObjectInstance.Value
	}
}

func (d *Document) querySelector() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		return nil
	}
}

func (d *Document) createTextNode() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		return d.createDummyObject(info.Context())
	}
}

func (d *Document) Inject(ctx *v8go.Context) error {
	iso := ctx.Isolate()
	global := ctx.Global()

	document := v8go.NewObjectTemplate(iso)
	if err := document.Set("createElement", v8go.NewFunctionTemplate(iso, d.createElement())); err != nil {
		return err
	}
	if err := document.Set("getElementById", v8go.NewFunctionTemplate(iso, d.getElementById())); err != nil {
		return err
	}
	if err := document.Set("querySelector", v8go.NewFunctionTemplate(iso, d.querySelector())); err != nil {
		return err
	}
	if err := document.Set("createTextNode", v8go.NewFunctionTemplate(iso, d.createTextNode())); err != nil {
		return err
	}

	documentInstance, err := document.NewInstance(ctx)
	if err != nil {
		return err
	}

	if err := global.Set("document", documentInstance); err != nil {
		return err
	}

	return nil
}
