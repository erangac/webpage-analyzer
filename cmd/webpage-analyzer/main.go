package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Serve static files from frontend/public
	fs := http.FileServer(http.Dir("frontend/public"))
	http.Handle("/", fs)

	// Basic API endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "webpage-analyzer"}`))
	})

	// API endpoint for analyzing webpages (placeholder)
	http.HandleFunc("/api/analyze", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Webpage analysis endpoint - coming soon!"}`))
	})

	log.Printf("Starting webpage analyzer server on port %s", *port)
	log.Printf("Frontend available at: http://localhost:%s", *port)
	log.Printf("Health check available at: http://localhost:%s/api/health", *port)
	
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 