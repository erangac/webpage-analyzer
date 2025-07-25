package http

import (
	"encoding/json"
	"net/http"
	"os"

	"webpage-analyzer/internal/analyzer"
)

// Handler handles HTTP requests for the webpage analyzer.
type Handler struct {
	analyzerService analyzer.Service
}

// NewHandler creates a new HTTP handler.
func NewHandler(analyzerService analyzer.Service) *Handler {
	return &Handler{
		analyzerService: analyzerService,
	}
}

// HealthCheck handles health check requests.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "healthy",
		"service": "webpage-analyzer",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AnalyzeWebpage handles webpage analysis requests.
func (h *Handler) AnalyzeWebpage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse request body.
	var req analyzer.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Analyze the webpage.
	analysis, err := h.analyzerService.AnalyzeWebpage(r.Context(), req)
	if err != nil {
		// Check if it's an AnalysisError.
		if analysisErr, ok := err.(*analyzer.AnalysisError); ok {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(analysisErr); err != nil {
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
				return
			}
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return analysis result.
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAnalysisStatus handles status requests.
func (h *Handler) GetAnalysisStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status, err := h.analyzerService.GetAnalysisStatus(r.Context())
	if err != nil {
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": status,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ServeOpenAPI serves the OpenAPI specification.
func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	openapiData, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		http.Error(w, "Failed to read OpenAPI spec", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(openapiData); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
