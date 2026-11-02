package sis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerSubscribers(t *testing.T) {
	config := newTestConfig()
	server := NewServer(config, NewStoreMem(config))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"get existing", http.MethodGet, "/subscribers/380670000001", "", http.StatusOK},
		{"get missing", http.MethodGet, "/subscribers/380679999999", "", http.StatusNotFound},
		{"bad length", http.MethodGet, "/subscribers/38067", "", http.StatusBadRequest},
		{"bad cc", http.MethodGet, "/subscribers/490670000001", "", http.StatusBadRequest},
		{"bad ndc", http.MethodGet, "/subscribers/380990000001", "", http.StatusBadRequest},
		{"put", http.MethodPut, "/subscribers/380679999999", `{"billing_type":1,"language_type":1,"operator_type":1}`, http.StatusOK},
		{"get after put", http.MethodGet, "/subscribers/380679999999", "", http.StatusOK},
		{"put missing field", http.MethodPut, "/subscribers/380679999999", `{"billing_type":1}`, http.StatusBadRequest},
		{"put negative", http.MethodPut, "/subscribers/380679999999", `{"billing_type":-1,"language_type":1,"operator_type":1}`, http.StatusBadRequest},
		{"put bad json", http.MethodPut, "/subscribers/380679999999", `{`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err := server.app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}
