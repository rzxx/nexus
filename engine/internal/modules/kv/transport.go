package kv

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *Module) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}

	item, found := m.store.Get(key)
	if !found {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item.Value)
}

func (m *Module) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
		TTL   int    `json:"ttl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	m.store.Set(req.Key, req.Value, req.TTL)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"success\":true}")
}
