package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"blog-platform/internal/application/service"
)

// RequestResponseLogger creates middleware for comprehensive request/response logging
func RequestResponseLogger(logger service.Logger) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Create a custom response writer to capture response body
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(res.Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: res.Writer}
			res.Writer = writer

			// Read and log request body for non-GET requests
			var reqBody []byte
			if req.Method != "GET" && req.Method != "DELETE" {
				if req.Body != nil {
					reqBody, _ = io.ReadAll(req.Body)
					req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
				}
			}

			// Log request
			ctx := req.Context()
			logger.Info(ctx, "HTTP request started",
				"method", req.Method,
				"uri", req.RequestURI,
				"remote_addr", c.RealIP(),
				"user_agent", req.UserAgent(),
				"content_length", req.ContentLength,
			)

			// Log request body for debugging (be careful with sensitive data)
			if len(reqBody) > 0 && len(reqBody) < 1024 { // Only log small bodies
				logger.Debug(ctx, "Request body", "body", string(reqBody))
			}

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			status := res.Status
			responseSize := res.Size

			logLevel := "info"
			if status >= 400 && status < 500 {
				logLevel = "warn"
			} else if status >= 500 {
				logLevel = "error"
			}

			logFields := []interface{}{
				"method", req.Method,
				"uri", req.RequestURI,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"response_size", responseSize,
				"remote_addr", c.RealIP(),
			}

			switch logLevel {
			case "warn":
				logger.Warn(ctx, "HTTP request completed with client error", logFields...)
			case "error":
				logger.Error(ctx, "HTTP request completed with server error", logFields...)
			default:
				logger.Info(ctx, "HTTP request completed", logFields...)
			}

			// Log response body for errors (be careful with sensitive data)
			if status >= 400 && resBody.Len() > 0 && resBody.Len() < 1024 {
				logger.Debug(ctx, "Response body", "body", resBody.String())
			}

			return err
		}
	})
}

// bodyDumpResponseWriter wraps the response writer to capture response body
type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, echo.NewHTTPError(http.StatusInternalServerError, "hijacking not supported")
}

// SecurityHeaders middleware adds security headers to all responses
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			
			// Security headers
			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("X-XSS-Protection", "1; mode=block")
			res.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			res.Header().Set("Content-Security-Policy", "default-src 'self'")
			
			// API-specific headers
			res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			res.Header().Set("Pragma", "no-cache")
			res.Header().Set("Expires", "0")
			
			return next(c)
		}
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			
			rid := req.Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = generateRequestID()
			}
			
			res.Header().Set(echo.HeaderXRequestID, rid)
			c.Set("request_id", rid)
			
			return next(c)
		}
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
