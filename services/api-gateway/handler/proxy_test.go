package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGatewayHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProxyHandler("http://mock-auth", "http://mock-prod", "http://mock-cart", "http://mock-order", "http://mock-pay")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.HealthCheck(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGatewayProxyForwarding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock downstream server
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be forwarded")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"mock": "response"}`))
	}))
	defer downstream.Close()

	h := NewProxyHandler(downstream.URL, downstream.URL, downstream.URL, downstream.URL, downstream.URL)

	r := gin.New()
	r.Any("/api/test/*path", h.ProxyTo(downstream.URL, "/api/test"))

	req := httptest.NewRequest("GET", "/api/test/data", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != `{"mock": "response"}` {
		t.Errorf("Expected mock response, got %s", w.Body.String())
	}
}

func TestGatewayServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Target unavailable port
	h := NewProxyHandler("http://localhost:59999", "http://localhost:59999", "http://localhost:59999", "http://localhost:59999", "http://localhost:59999")

	r := gin.New()
	r.Any("/api/offline/*path", h.ProxyTo("http://localhost:59999", "/api/offline"))

	req := httptest.NewRequest("GET", "/api/offline/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}
