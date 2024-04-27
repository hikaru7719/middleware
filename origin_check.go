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
	AllowSite      SecFetchSite
	Handler        http.Handler
	ErrorHandler   http.Handler
}

type originCheck struct {
	validateMethod []string
	allowOrigin    []string
	allowSite      SecFetchSite
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

func (oc *originCheck) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !slices.Contains(oc.validateMethod, req.Method) {
		oc.handler.ServeHTTP(w, req)
		return
	}

	origin := req.Header.Get(originHeader)
	if !slices.Contains(oc.allowOrigin, origin) {
		oc.errorHandler.ServeHTTP(w, req)
		return
	}

	secFetchSite := req.Header.Get(secFetchSiteHeader)
	if secFetchSite != "" && oc.allowSite.String() != secFetchSite {
		oc.errorHandler.ServeHTTP(w, req)
		return
	}

	oc.handler.ServeHTTP(w, req)
	return
}
