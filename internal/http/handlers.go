package http

import (
	"encoding/json"
	"net/http"
	"os"

	"webpage-analyzer/internal/analyzer"
)

// Handler handles HTTP requests for the webpage analyzer
type Handler struct {
	analyzerService analyzer.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(analyzerService analyzer.Service) *Handler {
	return &Handler{
		analyzerService: analyzerService,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ok",
		"service": "webpage-analyzer",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AnalyzeWebpage handles webpage analysis requests
func (h *Handler) AnalyzeWebpage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req analyzer.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	analysis, err := h.analyzerService.AnalyzeWebpage(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to analyze webpage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analysis)
}

// GetAnalysisStatus handles status requests
func (h *Handler) GetAnalysisStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.analyzerService.GetAnalysisStatus(r.Context())
	if err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ServeOpenAPI serves the OpenAPI specification
func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	// Read the OpenAPI specification file
	openapiData, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(openapiData)
} 