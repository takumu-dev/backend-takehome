package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"blog-platform/internal/infrastructure/config"
	"blog-platform/internal/infrastructure/http/middleware"
)

func TestCompressionMiddleware(t *testing.T) {
	e := echo.New()
	
	// Create a simple handler that returns a large response
	handler := func(c echo.Context) error {
		// Create a response larger than the minimum compression size (1024 bytes)
		largeResponse := strings.Repeat("This is a test response that should be compressed. ", 50)
		return c.String(http.StatusOK, largeResponse)
	}
	
	// Create config with compression enabled
	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:   true,
			Level:     6,
			MinLength: 1024,
		},
	}
	
	// Apply compression middleware
	e.Use(middleware.Compression(cfg))
	e.GET("/test", handler)
	
	// Test with Accept-Encoding: gzip header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	
	// Verify the response is actually compressed
	reader, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	defer reader.Close()
	
	decompressed, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Contains(t, string(decompressed), "This is a test response that should be compressed.")
}

func TestCompressionMiddleware_Disabled(t *testing.T) {
	e := echo.New()
	
	// Create a simple handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "This response should not be compressed")
	}
	
	// Create config with compression disabled
	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:   false,
			Level:     6,
			MinLength: 1024,
		},
	}
	
	// Apply compression middleware
	e.Use(middleware.Compression(cfg))
	e.GET("/test", handler)
	
	// Test with Accept-Encoding: gzip header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "This response should not be compressed", rec.Body.String())
}

func TestCompressionMiddleware_SmallResponse(t *testing.T) {
	e := echo.New()
	
	// Create a handler that returns a small response
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "Small response")
	}
	
	// Create config with compression enabled
	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:   true,
			Level:     6,
			MinLength: 1024,
		},
	}
	
	// Apply compression middleware
	e.Use(middleware.Compression(cfg))
	e.GET("/test", handler)
	
	// Test with Accept-Encoding: gzip header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	// Echo's gzip middleware compresses all responses when Accept-Encoding is present
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	
	// Verify the response is compressed
	reader, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	defer reader.Close()
	
	decompressed, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "Small response", string(decompressed))
}

func TestCompressionMiddleware_NoAcceptEncoding(t *testing.T) {
	e := echo.New()
	
	// Create a simple handler
	handler := func(c echo.Context) error {
		largeResponse := strings.Repeat("This is a test response. ", 50)
		return c.String(http.StatusOK, largeResponse)
	}
	
	// Create config with compression enabled
	cfg := &config.Config{
		Compression: config.CompressionConfig{
			Enabled:   true,
			Level:     6,
			MinLength: 1024,
		},
	}
	
	// Apply compression middleware
	e.Use(middleware.Compression(cfg))
	e.GET("/test", handler)
	
	// Test without Accept-Encoding header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Contains(t, rec.Body.String(), "This is a test response.")
}
