package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOriginCheck(t *testing.T) {
	tests := []struct {
		name         string
		config       OriginCheckConfig
		origin       string
		site         string
		expectStatus int
		expectBody   string
	}{
		{
			name: "same origin",
			config: OriginCheckConfig{
				AllowOrigin: []string{"http://example.com:8080"},
				AllowSite:   []SecFetchSite{SecFetchSiteSameOrigin},
			},
			origin:       "http://example.com:8080",
			site:         "same-origin",
			expectStatus: http.StatusOK,
		},
		{
			name: "same site",
			config: OriginCheckConfig{
				AllowOrigin: []string{"http://example.com:8080", "http://test.example.com:8080"},
				AllowSite:   []SecFetchSite{SecFetchSiteSameOrigin, SecFetchSiteSameSite},
			},
			origin:       "http://test.example.com:8080",
			site:         "same-site",
			expectStatus: http.StatusOK,
		},
		{
			name: "validate same origin",
			config: OriginCheckConfig{
				AllowOrigin: []string{"http://example.com:8080"},
				AllowSite:   []SecFetchSite{SecFetchSiteSameOrigin},
			},
			origin:       "http://test.example.com:8080",
			site:         "same-site",
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "validate same site",
			config: OriginCheckConfig{
				AllowOrigin: []string{"http://example.com:8080", "http://test.example.com:8080"},
				AllowSite:   []SecFetchSite{SecFetchSiteSameOrigin, SecFetchSiteSameSite},
			},
			origin:       "https://www.google.com",
			site:         "cross-site",
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "custom error handler",
			config: OriginCheckConfig{
				AllowOrigin: []string{"http://example.com:8080"},
				AllowSite:   []SecFetchSite{SecFetchSiteSameOrigin},
				ErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Error"))
				}),
			},
			origin:       "http://test.example.com:8080",
			site:         "same-site",
			expectStatus: http.StatusInternalServerError,
			expectBody:   "Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/test", nil)
			w := httptest.NewRecorder()

			if tc.origin != "" {
				r.Header.Add(originHeader, tc.origin)
			}
			if tc.site != "" {
				r.Header.Add(secFetchSiteHeader, tc.site)
			}

			handler := OriginCheck(tc.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(w, r)

			if tc.expectStatus != w.Code {
				t.Errorf("expect status code is %d, but got %d", tc.expectStatus, w.Code)
			}
			if w.Body.String() != "" {
				if tc.expectBody != w.Body.String() {
					t.Errorf("expect status code is %s, but got %s", tc.expectBody, w.Body.String())
				}
			}
		})
	}
}
