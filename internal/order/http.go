package order

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(svc *Service) http.Handler {
	r := chi.NewRouter()

	r.Post("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			UserID string `json:"userId"`
			Items  []Item `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		idemKey := r.Header.Get("Idempotency-Key")

		o, err := svc.Create(r.Context(), CreateOrderRequest{
			UserID:         body.UserID,
			Items:          body.Items,
			IdempotencyKey: idemKey,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(o)
	})

	r.Get("/api/v1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		o, err := svc.Get(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	})

	return r
}
