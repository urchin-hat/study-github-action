package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler := setupRoutes()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if res.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", res.Status)
	}
}

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryURL       string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "default world",
			queryURL:       "/hello",
			expectedBody:   "Hello, World!",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom name",
			queryURL:       "/hello?name=Gopher",
			expectedBody:   "Hello, Gopher!",
			expectedStatus: http.StatusOK,
		},
	}

	handler := setupRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.queryURL, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			body := strings.TrimSpace(rec.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body '%s', got '%s'", tt.expectedBody, body)
			}
		})
	}
}
