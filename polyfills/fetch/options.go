package fetch

import (
	"net/http"
	"net/url"
)

type UserAgentProvider interface {
	GetUserAgent(u *url.URL) string
}

type UserAgentProviderFunc func(u *url.URL) string

func (f UserAgentProviderFunc) GetUserAgent(u *url.URL) string {
	return f(u)
}

type Option interface {
	apply(ft *Fetcher)
}

type optionFunc func(ft *Fetcher)

func (f optionFunc) apply(ft *Fetcher) {
	f(ft)
}

func WithLocalHandler(handler http.Handler) Option {
	return optionFunc(func(ft *Fetcher) {
		ft.localHandler = handler
	})
}

func WithUserAgentProvider(provider UserAgentProvider) Option {
	return optionFunc(func(ft *Fetcher) {
		ft.userAgentProvider = provider
	})
}

func WithAddrLocal(addr string) Option {
	return optionFunc(func(ft *Fetcher) {
		ft.localAddress = addr
	})
}
