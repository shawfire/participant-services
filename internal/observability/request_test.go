package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := RequestID(r.Context())
			if id == "" {
				t.Error("request ID missing from context")
			}

			if id != w.Header().Get(RequestIDHeader) {
				t.Error("context and response request IDs differ")
			}

			w.WriteHeader(http.StatusOK)
		},
	))

	first := httptest.NewRecorder()
	handler.ServeHTTP(
		first,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	second := httptest.NewRecorder()
	handler.ServeHTTP(
		second,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	firstID := first.Header().Get(RequestIDHeader)
	secondID := second.Header().Get(RequestIDHeader)

	if len(firstID) != 32 || len(secondID) != 32 {
		t.Fatal("expected 32-character hexadecimal request IDs")
	}

	if firstID == secondID {
		t.Fatal("different requests received identical IDs")
	}

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatal("unexpected HTTP response status")
	}
}
