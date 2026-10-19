package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	c, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/onboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct {
			Domain     string `json:"domain"`
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Domain == "" || in.WebhookURL == "" {
			http.Error(w, "domain and webhook_url are required", http.StatusBadRequest)
			return
		}
		result, err := onboard(c, in.Domain, in.WebhookURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	log.Println("domainctl listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
