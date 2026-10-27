package health

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestReadinessInitiallyReady(t *testing.T) {
	readiness := NewReadiness()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	readiness.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", recorder.Code)
	}
}

func TestReadinessDuringShutdown(t *testing.T) {
	readiness := NewReadiness()
	readiness.SetReady(false)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	readiness.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected HTTP 503, got %d", recorder.Code)
	}
}

func TestReadinessConcurrentAccess(t *testing.T) {
	readiness := NewReadiness()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			readiness.SetReady(false)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodGet,
				"/readyz",
				nil,
			)

			readiness.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK &&
				recorder.Code != http.StatusServiceUnavailable {
				t.Errorf(
					"unexpected status: %d",
					recorder.Code,
				)
			}
		}()
	}

	wg.Wait()
}
