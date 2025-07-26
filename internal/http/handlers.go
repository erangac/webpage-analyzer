package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

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
		slog.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// writeError writes an error response with proper status code and message.
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	slog.Warn("HTTP error response", "status_code", statusCode, "message", message)
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
	start := time.Now()

	response := map[string]string{
		"status":  "healthy",
		"service": "webpage-analyzer",
	}
	h.writeJSON(w, http.StatusOK, response)

	slog.Info("Health check completed",
		"method", r.Method,
		"path", r.URL.Path,
		"duration", time.Since(start),
		"status_code", http.StatusOK,
	)
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
	start := time.Now()

	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for analyze endpoint",
			"method", r.Method,
			"path", r.URL.Path,
			"expected_method", http.MethodPost,
		)
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body.
	var req analyzer.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Failed to decode request body",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	slog.Info("Starting webpage analysis",
		"method", r.Method,
		"path", r.URL.Path,
		"url", req.URL,
	)

	// Analyze the webpage.
	analysis, err := h.analyzerService.AnalyzeWebpage(r.Context(), req)
	if err != nil {
		// Check if it's an AnalysisError and return it as JSON.
		if analysisErr, ok := err.(*analyzer.AnalysisError); ok {
			slog.Warn("Analysis failed with analysis error",
				"method", r.Method,
				"path", r.URL.Path,
				"url", req.URL,
				"error_type", "analysis_error",
				"status_code", analysisErr.StatusCode,
				"error_message", analysisErr.ErrorMessage,
				"duration", time.Since(start),
			)
			h.writeJSON(w, http.StatusBadRequest, analysisErr)
			return
		}
		// For other errors, return a generic error message.
		slog.Error("Analysis failed with internal error",
			"method", r.Method,
			"path", r.URL.Path,
			"url", req.URL,
			"error", err,
			"duration", time.Since(start),
		)
		h.writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Return analysis result.
	h.writeJSON(w, http.StatusOK, analysis)

	slog.Info("Webpage analysis completed successfully",
		"method", r.Method,
		"path", r.URL.Path,
		"url", req.URL,
		"status_code", http.StatusOK,
		"duration", time.Since(start),
		"has_login_form", analysis.HasLoginForm,
		"internal_links", analysis.InternalLinks,
		"external_links", analysis.ExternalLinks,
		"inaccessible_links", analysis.InaccessibleLinks,
		"headings_count", len(analysis.Headings),
	)
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
	start := time.Now()

	status, err := h.analyzerService.GetAnalysisStatus(r.Context())
	if err != nil {
		slog.Error("Failed to get analysis status",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
			"duration", time.Since(start),
		)
		h.writeError(w, http.StatusInternalServerError, "Failed to get status")
		return
	}

	response := map[string]string{
		"status": status,
	}
	h.writeJSON(w, http.StatusOK, response)

	slog.Info("Status request completed",
		"method", r.Method,
		"path", r.URL.Path,
		"status_code", http.StatusOK,
		"duration", time.Since(start),
	)
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
	start := time.Now()

	w.Header().Set("Content-Type", "application/yaml")
	// Serve the dynamically generated OpenAPI spec
	openapiData, err := os.ReadFile(openAPIFilePath)
	if err != nil {
		slog.Error("Failed to read OpenAPI spec file",
			"method", r.Method,
			"path", r.URL.Path,
			"file_path", openAPIFilePath,
			"error", err,
			"duration", time.Since(start),
		)
		h.writeError(w, http.StatusInternalServerError, "Failed to read OpenAPI spec")
		return
	}
	if _, err := w.Write(openapiData); err != nil {
		slog.Error("Failed to write OpenAPI spec response",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
			"duration", time.Since(start),
		)
		h.writeError(w, http.StatusInternalServerError, "Failed to write response")
		return
	}

	slog.Info("OpenAPI spec served successfully",
		"method", r.Method,
		"path", r.URL.Path,
		"status_code", http.StatusOK,
		"duration", time.Since(start),
		"file_size", len(openapiData),
	)
}
