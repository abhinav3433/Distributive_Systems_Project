package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	authURL    string
	productURL string
	cartURL    string
	orderURL   string
	paymentURL string
	client     *http.Client
}

func NewProxyHandler(authURL, productURL, cartURL, orderURL, paymentURL string) *ProxyHandler {
	return &ProxyHandler{
		authURL:    strings.TrimRight(authURL, "/"),
		productURL: strings.TrimRight(productURL, "/"),
		cartURL:    strings.TrimRight(cartURL, "/"),
		orderURL:   strings.TrimRight(orderURL, "/"),
		paymentURL: strings.TrimRight(paymentURL, "/"),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *ProxyHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "UP",
		"service": "api-gateway",
	})
}

func (h *ProxyHandler) ProxyTo(targetBaseURL, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqPath := c.Request.URL.Path
		forwardPath := strings.TrimPrefix(reqPath, stripPrefix)
		if !strings.HasPrefix(forwardPath, "/") {
			forwardPath = "/" + forwardPath
		}

		targetURL := targetBaseURL + forwardPath
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		outReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upstream request: " + err.Error()})
			return
		}

		// Forward headers
		for k, vv := range c.Request.Header {
			for _, v := range vv {
				outReq.Header.Add(k, v)
			}
		}

		// Append X-Forwarded-For
		clientIP := c.ClientIP()
		if xff := outReq.Header.Get("X-Forwarded-For"); xff != "" {
			outReq.Header.Set("X-Forwarded-For", xff+", "+clientIP)
		} else {
			outReq.Header.Set("X-Forwarded-For", clientIP)
		}

		resp, err := h.client.Do(outReq)
		if err != nil {
			log.Printf("[GATEWAY] Error proxying %s %s -> %s: %v", c.Request.Method, reqPath, targetURL, err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service unavailable",
				"details": fmt.Sprintf("Downstream service at %s could not be reached", targetBaseURL),
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Add(k, v)
			}
		}

		c.Status(resp.StatusCode)
		_, _ = io.Copy(c.Writer, resp.Body)

		log.Printf("[GATEWAY] %s %s -> %s [%d] in %v", c.Request.Method, reqPath, targetURL, resp.StatusCode, time.Since(start))
	}
}

// Proxies for individual microservices
func (h *ProxyHandler) AuthProxy() gin.HandlerFunc {
	return h.ProxyTo(h.authURL, "/api/auth")
}

func (h *ProxyHandler) ProductProxy() gin.HandlerFunc {
	return h.ProxyTo(h.productURL, "/api")
}

func (h *ProxyHandler) CartProxy() gin.HandlerFunc {
	return h.ProxyTo(h.cartURL, "/api")
}

func (h *ProxyHandler) OrderProxy() gin.HandlerFunc {
	return h.ProxyTo(h.orderURL, "/api")
}

func (h *ProxyHandler) PaymentProxy() gin.HandlerFunc {
	return h.ProxyTo(h.paymentURL, "/api")
}
