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
// @Summary Health check
// @Description Check if the service is running and healthy
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
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
// @Summary Analyze webpage
// @Description Analyze a webpage and return comprehensive information including HTML version, page title, headings structure, link analysis, and login form detection
// @Tags Analysis
// @Accept json
// @Produce json
// @Param request body analyzer.AnalysisRequest true "Analysis request"
// @Success 200 {object} analyzer.WebpageAnalysis
// @Failure 400 {object} analyzer.AnalysisError
// @Failure 500 {object} map[string]string
// @Router /api/analyze [post]
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
// @Summary Get service status
// @Description Get the current status and capabilities of the analysis service
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/status [get]
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
// @Summary Get OpenAPI specification
// @Description Retrieve the OpenAPI specification for this API
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {string} string "OpenAPI specification"
// @Failure 500 {object} map[string]string
// @Router /api/openapi [get]
func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	// Serve the dynamically generated OpenAPI spec
	openapiData, err := os.ReadFile("api/swagger.yaml")
	if err != nil {
		http.Error(w, "Failed to read OpenAPI spec", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(openapiData); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
