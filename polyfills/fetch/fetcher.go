package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"rogchap.com/v8go"
)

const (
	UserAgentLocal   = "<local>"
	LocalAddress     = "0.0.0.0:0"
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
)

var defaultLocalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
})

var defaultUserAgentProvider = UserAgentProviderFunc(func(u *url.URL) string {
	if !u.IsAbs() {
		return UserAgentLocal
	}

	return DefaultUserAgent
})

type Fetcher struct {
	localHandler      http.Handler
	userAgentProvider UserAgentProvider
	localAddress      string
}

func New(opt ...Option) *Fetcher {
	fetcher := &Fetcher{
		localHandler:      defaultLocalHandler,
		userAgentProvider: defaultUserAgentProvider,
		localAddress:      LocalAddress,
	}

	for _, o := range opt {
		o.apply(fetcher)
	}

	return fetcher
}

func (f *Fetcher) GetLocalHandler() http.Handler {
	return f.localHandler
}

func (f *Fetcher) fetch() v8go.FunctionCallback {
	return func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		args := info.Args()

		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			if len(args) <= 0 {
				err := errors.New("1 argument required, but only 0 present")
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			var reqInit RequestInit
			if len(args) > 1 {
				str, err := v8go.JSONStringify(ctx, args[1])
				if err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}

				reader := strings.NewReader(str)
				if err := json.NewDecoder(reader).Decode(&reqInit); err != nil {
					resolver.Reject(newErrorValue(ctx, err))
					return
				}
			}

			r, err := f.initRequest(args[0].String(), reqInit)
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			var res *Response

			// do local request
			if !r.URL.IsAbs() {
				res, err = f.fetchLocal(r)
			} else {
				res, err = f.fetchRemote(r)
			}
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			resObj, err := newResponseObject(ctx, res)
			if err != nil {
				resolver.Reject(newErrorValue(ctx, err))
				return
			}

			resolver.Resolve(resObj)
		}()

		return resolver.GetPromise().Value
	}
}

func (f *Fetcher) initRequest(reqUrl string, reqInit RequestInit) (*Request, error) {
	u, err := ParseRequestURL(reqUrl)
	if err != nil {
		return nil, err
	}

	req := &Request{
		URL:  u,
		Body: reqInit.Body,
		Header: http.Header{
			"Accept":     []string{"*/*"},
			"Connection": []string{"close"},
		},
	}

	var ua string
	if f.userAgentProvider != nil {
		ua = f.userAgentProvider.GetUserAgent(u)
	} else {
		ua = defaultUserAgentProvider(u)
	}

	req.Header.Set("User-Agent", ua)

	// url has no scheme, it's a local request
	if !u.IsAbs() {
		req.RemoteAddr = f.localAddress
	}

	for h, v := range reqInit.Headers {
		headerName := http.CanonicalHeaderKey(h)
		req.Header.Set(headerName, v)
	}

	if reqInit.Method != "" {
		req.Method = strings.ToUpper(reqInit.Method)
	} else {
		req.Method = "GET"
	}

	switch r := strings.ToLower(reqInit.Redirect); r {
	case "error", "follow", "manual":
		req.Redirect = r
	case "":
		req.Redirect = RequestRedirectFollow
	default:
		return nil, fmt.Errorf("unsupported redirect: %s", reqInit.Redirect)
	}

	return req, nil
}

func (f *Fetcher) fetchLocal(r *Request) (*Response, error) {
	if f.localHandler == nil {
		return nil, errors.New("no local handler present")
	}

	var body io.Reader
	if r.Method != "GET" {
		body = strings.NewReader(r.Body)
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.RemoteAddr = r.RemoteAddr
	req.Header = r.Header

	rcd := httptest.NewRecorder()

	f.localHandler.ServeHTTP(rcd, req)

	return HandleHttpResponse(rcd.Result(), r.URL.String(), false)
}

func (f *Fetcher) fetchRemote(r *Request) (*Response, error) {
	var body io.Reader
	if r.Method != "GET" {
		body = strings.NewReader(r.Body)
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	redirected := false
	client := &http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			switch r.Redirect {
			case RequestRedirectError:
				return errors.New("redirects are not allowed")
			default:
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
			}

			redirected = true
			return nil
		},
		Timeout: 20 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return HandleHttpResponse(res, r.URL.String(), redirected)
}

func newResponseObject(ctx *v8go.Context, res *Response) (*v8go.Object, error) {
	iso := ctx.Isolate()

	headers, err := newHeadersObject(ctx, res.Header)
	if err != nil {
		return nil, err
	}

	textFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()
		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			v, _ := v8go.NewValue(iso, res.Body)
			resolver.Resolve(v)
		}()

		return resolver.GetPromise().Value
	})
	if err != nil {
		return nil, err
	}

	jsonFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := info.Context()

		resolver, _ := v8go.NewPromiseResolver(ctx)

		go func() {
			val, err := v8go.JSONParse(ctx, res.Body)
			if err != nil {
				rejectVal, _ := v8go.NewValue(iso, err.Error())
				resolver.Reject(rejectVal)
				return
			}

			resolver.Resolve(val)
		}()

		return resolver.GetPromise().Value
	})
	if err != nil {
		return nil, err
	}

	resTmp := v8go.NewObjectTemplate(iso)

	for _, f := range []struct {
		Name string
		Tmp  interface{}
	}{
		{Name: "text", Tmp: textFnTmp},
		{Name: "json", Tmp: jsonFnTmp},
	} {
		if err := resTmp.Set(f.Name, f.Tmp, v8go.ReadOnly); err != nil {
			return nil, err
		}
	}

	resObj, err := resTmp.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range []struct {
		Key string
		Val interface{}
	}{
		{Key: "headers", Val: headers},
		{Key: "ok", Val: res.OK},
		{Key: "redirected", Val: res.Redirected},
		{Key: "status", Val: res.Status},
		{Key: "statusText", Val: res.StatusText},
		{Key: "url", Val: res.URL},
		{Key: "body", Val: res.Body},
	} {
		if err := resObj.Set(v.Key, v.Val); err != nil {
			return nil, err
		}
	}

	return resObj, nil
}

func newHeadersObject(ctx *v8go.Context, h http.Header) (*v8go.Object, error) {
	iso := ctx.Isolate()

	// https://developer.mozilla.org/en-US/docs/Web/API/Headers/get
	getFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) <= 0 {
			// TODO: this should return an error, but v8go not supported now
			val, _ := v8go.NewValue(iso, "")
			return val
		}

		key := http.CanonicalHeaderKey(args[0].String())
		val, _ := v8go.NewValue(iso, h.Get(key))
		return val
	})

	// https://developer.mozilla.org/en-US/docs/Web/API/Headers/has
	hasFnTmp := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) <= 0 {
			val, _ := v8go.NewValue(iso, false)
			return val
		}
		key := http.CanonicalHeaderKey(args[0].String())

		val, _ := v8go.NewValue(iso, h.Get(key) != "")
		return val
	})

	// create a header template,
	// TODO: if v8go supports Map in the future, change this to a Map Object
	headersTmp := v8go.NewObjectTemplate(iso)

	for _, f := range []struct {
		Name string
		Tmp  interface{}
	}{
		{Name: "get", Tmp: getFnTmp},
		{Name: "has", Tmp: hasFnTmp},
	} {
		if err := headersTmp.Set(f.Name, f.Tmp, v8go.ReadOnly); err != nil {
			return nil, err
		}
	}

	headers, err := headersTmp.NewInstance(ctx)
	if err != nil {
		return nil, err
	}

	for k, v := range h {
		var vv string
		if len(v) > 0 {
			// get the first element, like http.Header.Get
			vv = v[0]
		}

		if err := headers.Set(k, vv); err != nil {
			return nil, err
		}
	}

	return headers, nil
}

func newErrorValue(ctx *v8go.Context, err error) *v8go.Value {
	iso := ctx.Isolate()
	e, _ := v8go.NewValue(iso, fmt.Sprintf("fetch: %v", err))
	return e
}
