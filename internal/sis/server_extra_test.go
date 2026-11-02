package sis

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubStore lets tests control store errors and latency.
type stubStore struct {
	err   error
	delay time.Duration
}

func (s *stubStore) Get(ctx context.Context, _ int64) (*Subscriber, error) {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return nil, s.err
}
func (s *stubStore) Set(ctx context.Context, _ *Subscriber) error {
	_, err := s.Get(ctx, 0)
	return err
}
func (s *stubStore) Ping(context.Context) error { return s.err }
func (s *stubStore) Close()                     {}

func doRequest(t *testing.T, server *Server, method, path string) (int, string) {
	t.Helper()
	resp, err := server.app.Test(httptest.NewRequest(method, path, nil), -1)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestServerStoreErrors(t *testing.T) {
	config := newTestConfig()
	config.DbTimeout = 50 * time.Millisecond

	t.Run("db error", func(t *testing.T) {
		server := NewServer(config, &stubStore{err: errors.New("boom")})
		code, body := doRequest(t, server, http.MethodGet, "/subscribers/380670000001")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Contains(t, body, `"error_id":10`)

		code, _ = doRequest(t, server, http.MethodGet, "/ready")
		assert.Equal(t, http.StatusServiceUnavailable, code)
	})

	t.Run("db timeout", func(t *testing.T) {
		server := NewServer(config, &stubStore{delay: time.Second})
		code, body := doRequest(t, server, http.MethodGet, "/subscribers/380670000001")
		assert.Equal(t, http.StatusGatewayTimeout, code)
		assert.Contains(t, body, `"error_id":11`)
	})
}

func TestServerReadyAndMetrics(t *testing.T) {
	config := newTestConfig()
	server := NewServer(config, NewStoreMem(config))

	code, _ := doRequest(t, server, http.MethodGet, "/ready")
	assert.Equal(t, http.StatusOK, code)

	doRequest(t, server, http.MethodGet, "/subscribers/380670000001")
	doRequest(t, server, http.MethodGet, "/subscribers/380670000002")
	doRequest(t, server, http.MethodGet, "/subscribers/380679999999")
	doRequest(t, server, http.MethodGet, "/subscribers/bad")

	code, body := doRequest(t, server, http.MethodGet, "/metrics")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "# TYPE sis_requests_total counter")
	assert.Contains(t, body, `sis_requests_total{method="GET",code="200"} 2`)
	assert.Contains(t, body, `sis_requests_total{method="GET",code="404"} 1`)
	assert.Contains(t, body, `sis_requests_total{method="GET",code="400"} 1`)
	assert.Contains(t, body, "subscribers_total 4")
}

func TestParseMsisdn(t *testing.T) {
	config := newTestConfig()
	config.CC = "1"
	config.NDCS = []int{212, 5}
	config.MsisdnLength = 11
	server := NewServer(config, &stubStore{})

	tests := []struct {
		msisdn string
		errID  byte
	}{
		{"12125550100", 0},
		{"15005550100", 0},
		{"13105550100", errUnsupportedNDC},
		{"42125550100", errUnsupportedCC},
		{"1212555010x", errMsisdnFormat},
		{"121255501", errMsisdnFormat},
	}
	for _, tt := range tests {
		t.Run(tt.msisdn, func(t *testing.T) {
			_, apiErr := server.parseMsisdn(tt.msisdn)
			if tt.errID == 0 {
				assert.Nil(t, apiErr)
				return
			}
			require.NotNil(t, apiErr)
			assert.Equal(t, tt.errID, apiErr.ErrorID, apiErr.ErrorMsg)
		})
	}
}
