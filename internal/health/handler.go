package health

import (
	"net/http"
	"sync/atomic"
)

type Readiness struct {
	ready atomic.Bool
}

func NewReadiness() *Readiness {
	r := &Readiness{}
	r.ready.Store(true)
	return r
}

func (r *Readiness) SetReady(ready bool) {
	r.ready.Store(ready)
}

func (r *Readiness) ServeHTTP(
	w http.ResponseWriter,
	req *http.Request,
) {
	if !r.ready.Load() {
		http.Error(
			w,
			"not ready",
			http.StatusServiceUnavailable,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready\n"))
}
