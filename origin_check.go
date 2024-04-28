package middleware

import (
	"net/http"
	"slices"
)

type SecFetchSite string

var (
	SecFetchSiteCrossSite  SecFetchSite = "cross-site"
	SecFetchSiteSameOrigin SecFetchSite = "same-origin"
	SecFetchSiteSameSite   SecFetchSite = "same-site"
	SecFetchSiteNone       SecFetchSite = "none"
)

func (s SecFetchSite) String() string {
	return string(s)
}

type OriginCheckConfig struct {
	ValidateMethod []string
	AllowOrigin    []string
	AllowSite      []SecFetchSite
	Handler        http.Handler
	ErrorHandler   http.Handler
}

type originCheck struct {
	validateMethod []string
	allowOrigin    []string
	allowSite      []SecFetchSite
	errorHandler   http.Handler
	handler        http.Handler
}

var (
	originHeader       = "origin"
	secFetchSiteHeader = "sec-fetch-site"
)

func OriginCheckWithConfig(originCheckConfig OriginCheckConfig) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &originCheck{
			validateMethod: originCheckConfig.ValidateMethod,
			allowOrigin:    originCheckConfig.AllowOrigin,
			allowSite:      originCheckConfig.AllowSite,
			errorHandler:   originCheckConfig.ErrorHandler,
			handler:        handler,
		}
	}
}

func (oc *originCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !slices.Contains(oc.validateMethod, r.Method) {
		oc.handler.ServeHTTP(w, r)
		return
	}

	origin := r.Header.Get(originHeader)
	if !slices.Contains(oc.allowOrigin, origin) {
		oc.errorHandler.ServeHTTP(w, r)
		return
	}

	secFetchSite := r.Header.Get(secFetchSiteHeader)
	if secFetchSite != "" && !slices.Contains(oc.allowSite, SecFetchSite(secFetchSite)) {
		oc.errorHandler.ServeHTTP(w, r)
		return
	}

	oc.handler.ServeHTTP(w, r)
	return
}
