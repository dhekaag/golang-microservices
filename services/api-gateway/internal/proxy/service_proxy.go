package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/config"
	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
)

type ServiceProxy struct {
	services map[string]*httputil.ReverseProxy
	config   *config.ServicesConfig
}

func NewServiceProxy(config *config.ServicesConfig) *ServiceProxy {
	services := make(map[string]*httputil.ReverseProxy)

	// User service proxy
	if userURL, err := url.Parse(config.UserService); err == nil {
		services["user"] = createReverseProxy(userURL, "user-service")
	} else {
		log.Printf("Failed to parse user service URL: %v", err)
	}

	// Product service proxy
	if productURL, err := url.Parse(config.ProductService); err == nil {
		services["product"] = createReverseProxy(productURL, "product-service")
	} else {
		log.Printf("Failed to parse product service URL: %v", err)
	}

	// Order service proxy
	if orderURL, err := url.Parse(config.OrderService); err == nil {
		services["order"] = createReverseProxy(orderURL, "order-service")
	} else {
		log.Printf("Failed to parse order service URL: %v", err)
	}

	return &ServiceProxy{
		services: services,
		config:   config,
	}
}

func createReverseProxy(target *url.URL, serviceName string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom director to modify requests
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 🔑 ENHANCED: Forward context headers
		if requestID := req.Header.Get("X-Request-ID"); requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}

		if correlationID := req.Header.Get("X-Correlation-ID"); correlationID != "" {
			req.Header.Set("X-Correlation-ID", correlationID)
		}

		if userID := req.Header.Get("X-User-ID"); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}

		// Add service identification headers
		req.Header.Set("X-Forwarded-By", "api-gateway")
		req.Header.Set("X-Target-Service", serviceName)
		req.Header.Set("User-Agent", "API-Gateway/1.0")

		// Remove sensitive headers that shouldn't be forwarded
		req.Header.Del("Cookie")
		req.Header.Del("Authorization")
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		_ = r.Context()
		log.Printf("❌ Proxy error for %s: %v", serviceName, err)

		utils.SendError(w, http.StatusBadGateway, fmt.Sprintf("Service %s is currently unavailable", serviceName))
	}

	// Custom modify response
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Forward response headers
		if requestID := resp.Header.Get("X-Request-ID"); requestID != "" {
			resp.Header.Set("X-Request-ID", requestID)
		}

		if correlationID := resp.Header.Get("X-Correlation-ID"); correlationID != "" {
			resp.Header.Set("X-Correlation-ID", correlationID)
		}

		// Add proxy headers
		resp.Header.Set("X-Proxied-By", "api-gateway")
		resp.Header.Set("X-Service-Name", serviceName)

		return nil
	}

	return proxy
}

func (sp *ServiceProxy) ProxyToService(serviceName string, w http.ResponseWriter, r *http.Request) {
	proxy, exists := sp.services[serviceName]
	if !exists {
		utils.SendError(w, http.StatusNotFound, fmt.Sprintf("Service %s not found", serviceName))
		return
	}

	// Add request tracing
	log.Printf("Proxying request to %s: %s %s", serviceName, r.Method, r.URL.Path)

	proxy.ServeHTTP(w, r)
}

func (sp *ServiceProxy) IsServiceHealthy(serviceName string) bool {
	var serviceURL string

	switch serviceName {
	case "user":
		serviceURL = sp.config.UserService
	case "product":
		serviceURL = sp.config.ProductService
	case "order":
		serviceURL = sp.config.OrderService
	default:
		return false
	}

	resp, err := http.Get(serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
