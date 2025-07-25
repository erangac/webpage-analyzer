package main

import (
	"flag"
	"log"
	"net/http"

	"webpage-analyzer/internal/analyzer"
	httphandler "webpage-analyzer/internal/http"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize services
	analyzerService := analyzer.NewService()

	// Initialize handlers
	handler := httphandler.NewHandler(analyzerService)

	// Serve static files from frontend/public
	fs := http.FileServer(http.Dir("frontend/public"))
	http.Handle("/", fs)

	// API routes
	http.HandleFunc("/api/health", handler.HealthCheck)
	http.HandleFunc("/api/analyze", handler.AnalyzeWebpage)
	http.HandleFunc("/api/status", handler.GetAnalysisStatus)
	
	// API Documentation routes
	http.HandleFunc("/api/openapi", handler.ServeOpenAPI)
	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/public/docs.html")
	})

	log.Printf("Starting webpage analyzer server on port %s", *port)
	log.Printf("Frontend available at: http://localhost:%s", *port)
	log.Printf("API Documentation available at: http://localhost:%s/docs", *port)
	log.Printf("Health check available at: http://localhost:%s/api/health", *port)
	log.Printf("Analysis endpoint available at: http://localhost:%s/api/analyze", *port)
	log.Printf("Status endpoint available at: http://localhost:%s/api/status", *port)
	log.Printf("OpenAPI spec available at: http://localhost:%s/api/openapi", *port)
	
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 