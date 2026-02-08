package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	// Identity Service mock
	mux.HandleFunc("/identity/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    7564,
			"name":  "Anupam C",
			"email": "anupam.choudhury@abc.xyz",
			"roles": []string{}, // toggle to []string{} for standard
		})
	})

	// Availability Service mock
	mux.HandleFunc("/availability/availabilityinfo/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 3 {
			http.Error(w, "bad path", http.StatusNotFound)
			return
		}
		videoID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || videoID <= 0 {
			http.Error(w, "invalid video_id", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"video_id": videoID,
			"availability_window": map[string]string{
				"from": "2025-11-05",
				"to":   "2026-05-04",
			},
		})
	})

	addr := "127.0.0.1:9001"
	log.Printf("mock-upstreams listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
