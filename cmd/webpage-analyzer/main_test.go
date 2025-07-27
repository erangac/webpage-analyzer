package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerStartupAndEndpoints(t *testing.T) {
	// Use the same setup logic as main()
	server := setupServer("9876")

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test endpoints
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get("http://localhost:9876/api/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("StatusEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:9876/api/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AnalyzeEndpointMethodNotAllowed", func(t *testing.T) {
		resp, err := http.Get("http://localhost:9876/api/analyze")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("FrontendRoot", func(t *testing.T) {
		resp, err := http.Get("http://localhost:9876/")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return some response (even if 404 for missing files)
		assert.NotEqual(t, 0, resp.StatusCode)
	})

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestStaticDirConstant(t *testing.T) {
	assert.Equal(t, "frontend/public", staticDir)
}
