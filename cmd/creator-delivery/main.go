package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"creator-delivery/delivery"
)

func main() {
	client, err := delivery.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/deliver", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
			return
		}
		var input delivery.Request
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		result, err := delivery.Deliver(r.Context(), client, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("creator delivery listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
