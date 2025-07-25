package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"webpage-analyzer/internal/analyzer"
	httphandler "webpage-analyzer/internal/http"
)

const (
	staticDir = "frontend/public"
)

func registerRoutes(handler *httphandler.Handler) {
	// Serve static files from frontend/public.
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fs)

	// API routes.
	http.HandleFunc("/api/health", handler.HealthCheck)
	http.HandleFunc("/api/analyze", handler.AnalyzeWebpage)
	http.HandleFunc("/api/status", handler.GetAnalysisStatus)

	// API Documentation routes.
	http.HandleFunc("/api/openapi", handler.ServeOpenAPI)
	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/docs.html")
	})
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize services.
	analyzerService := analyzer.NewService()

	// Initialize handlers.
	handler := httphandler.NewHandler(analyzerService)

	// Register all routes.
	registerRoutes(handler)

	log.Printf("Starting webpage analyzer server on port %s", *port)
	
	// Log available endpoints
	endpoints := []struct {
		name string
		path string
	}{
		{"Frontend", "/"},
		{"API Documentation", "/docs"},
		{"Health check", "/api/health"},
		{"Analysis endpoint", "/api/analyze"},
		{"Status endpoint", "/api/status"},
		{"OpenAPI spec", "/api/openapi"},
	}

	for _, endpoint := range endpoints {
		log.Printf("%s available at: http://localhost:%s%s", endpoint.name, *port, endpoint.path)
	}

	// Create server with timeout configuration.
	server := &http.Server{
		Addr:         ":" + *port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
