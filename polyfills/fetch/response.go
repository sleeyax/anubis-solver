package fetch

import (
	"io/ioutil"
	"net/http"
)

type Response struct {
	Header     http.Header
	Status     int32
	StatusText string
	OK         bool
	Redirected bool
	URL        string
	Body       string
}

func HandleHttpResponse(res *http.Response, url string, redirected bool) (*Response, error) {
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Header:     res.Header,
		Status:     int32(res.StatusCode), // int type is not support by v8go
		StatusText: res.Status,
		OK:         res.StatusCode >= 200 && res.StatusCode < 300,
		Redirected: redirected,
		URL:        url,
		Body:       string(resBody),
	}, nil
}
