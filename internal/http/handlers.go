package http

import (
	"encoding/json"
	"net/http"
	"os"

	"webpage-analyzer/internal/analyzer"
)

const (
	openAPIFilePath = "api/swagger.yaml"
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

// writeJSON writes a JSON response with proper headers and error handling.
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// writeError writes an error response with proper status code and message.
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
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
	response := map[string]string{
		"status":  "healthy",
		"service": "webpage-analyzer",
	}
	h.writeJSON(w, http.StatusOK, response)
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
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body.
	var req analyzer.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Analyze the webpage.
	analysis, err := h.analyzerService.AnalyzeWebpage(r.Context(), req)
	if err != nil {
		// Check if it's an AnalysisError and return it as JSON.
		if analysisErr, ok := err.(*analyzer.AnalysisError); ok {
			h.writeJSON(w, http.StatusBadRequest, analysisErr)
			return
		}
		// For other errors, return a generic error message.
		h.writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Return analysis result.
	h.writeJSON(w, http.StatusOK, analysis)
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
	status, err := h.analyzerService.GetAnalysisStatus(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get status")
		return
	}

	response := map[string]string{
		"status": status,
	}
	h.writeJSON(w, http.StatusOK, response)
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
	openapiData, err := os.ReadFile(openAPIFilePath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to read OpenAPI spec")
		return
	}
	if _, err := w.Write(openapiData); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to write response")
		return
	}
}
