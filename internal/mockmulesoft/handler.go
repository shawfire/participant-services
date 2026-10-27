package mockmulesoft

import (
	"encoding/json"
	"net/http"

	"github.com/shawfire/participant-services/internal/participant"
)

// NewHandler simulates a MuleSoft participant lookup API.
// All data is synthetic.
func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/participants/{id}", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		id := r.PathValue("id")

		participants := map[string]participant.Participant{
			"P001": {
				ID:        "P001",
				FirstName: "Alex",
				LastName:  "Taylor",
			},
			"P002": {
				ID:        "P002",
				FirstName: "Jordan",
				LastName:  "Lee",
			},
		}

		p, found := participants[id]
		if !found {
			http.Error(w, "participant not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(p); err != nil {
			return
		}
	})

	return mux
}
