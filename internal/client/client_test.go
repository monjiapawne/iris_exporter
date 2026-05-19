package client

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testResponse struct {
	Value int `json:"value"`
}

func TestAPICall(t *testing.T) {
	testCases := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
		want    testResponse
	}{
		{
			name: "successful response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"value": 42}`))
			},
			want: testResponse{Value: 42},
		},
		{
			name: "non-2xx status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr: true,
		},
		{
			name: "sets authorization header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Write([]byte(`{}`))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			c := &Client{
				Host:       srv.URL,
				APIKey:     "test-key",
				httpClient: srv.Client(),
				logger:     slog.Default(),
			}

			got, err := APICall[testResponse](c, "/test")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}
