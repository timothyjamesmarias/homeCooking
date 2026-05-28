package router

import (
	"encoding/json"
	"net/http"

	"home-cooking.timothymarias.com/internal/store"
)

func New(db *store.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Domain routes will be registered here as they're built out:
	// registerShoppingRoutes(mux, db)
	// registerRecipeRoutes(mux, db)
	// registerPantryRoutes(mux, db)

	return mux
}
