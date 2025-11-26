// ABOUTME: HTTP server with routing and middleware for the mock BOSH Director.
// ABOUTME: Handles TLS, authentication, and request routing.

package mockbosh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port     int
	Username string
	Password string
	UseTLS   bool
	Speed    float64
	Debug    bool
}

// DefaultServerConfig returns default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:     25555,
		Username: "admin",
		Password: "admin",
		UseTLS:   true,
		Speed:    1.0,
		Debug:    false,
	}
}

// Server is the mock BOSH Director HTTP server.
type Server struct {
	config     ServerConfig
	state      *State
	simulator  *TaskSimulator
	handlers   *Handlers
	httpServer *http.Server
}

// NewServer creates a new mock BOSH Director server.
func NewServer(config ServerConfig) *Server {
	state := NewState()
	simulator := NewTaskSimulator(state, config.Speed, config.Debug)
	handlers := NewHandlers(state, simulator, config.Username, config.Password)

	return &Server{
		config:    config,
		state:     state,
		simulator: simulator,
		handlers:  handlers,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	addr := fmt.Sprintf(":%d", s.config.Port)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.loggingMiddleware(s.authMiddleware(mux)),
	}

	protocol := "http"
	if s.config.UseTLS {
		protocol = "https"
		tlsConfig, err := s.generateTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to generate TLS config: %w", err)
		}
		s.httpServer.TLSConfig = tlsConfig
	}

	log.Printf("Mock BOSH Director starting on %s://localhost%s", protocol, addr)
	log.Printf("Credentials: %s / %s", s.config.Username, s.config.Password)
	log.Printf("Simulation speed: %.1fx", s.config.Speed)

	if s.config.UseTLS {
		return s.httpServer.ListenAndServeTLS("", "")
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registers all API routes.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/info", s.handlers.HandleInfo)
	mux.HandleFunc("/deployments", s.routeDeployments)
	mux.HandleFunc("/deployments/", s.routeDeployments)
	mux.HandleFunc("/tasks", s.routeTasks)
	mux.HandleFunc("/tasks/", s.routeTasks)
	mux.HandleFunc("/stemcells", s.handlers.HandleStemcells)
	mux.HandleFunc("/releases", s.handlers.HandleReleases)
	mux.HandleFunc("/configs", s.handlers.HandleConfigs)
	mux.HandleFunc("/locks", s.handlers.HandleLocks)
}

// routeDeployments routes deployment-related requests.
func (s *Server) routeDeployments(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/deployments" {
		s.handlers.HandleDeployments(w, r)
		return
	}

	parts := strings.Split(strings.TrimPrefix(path, "/deployments/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "deployment name required")
		return
	}

	deployment := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			d, err := s.state.GetDeployment(deployment)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, d)
		case http.MethodDelete:
			s.handlers.HandleDeleteDeployment(w, r, deployment)
		case http.MethodPut:
			if r.URL.Query().Get("state") == "recreate" {
				s.handlers.HandleDeploymentRecreate(w, r, deployment)
			} else {
				writeError(w, http.StatusBadRequest, "unknown operation")
			}
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 && parts[1] == "vms" {
		s.handlers.HandleDeploymentVMs(w, r, deployment)
		return
	}

	if len(parts) == 2 && parts[1] == "instances" {
		s.handlers.HandleDeploymentInstances(w, r, deployment)
		return
	}

	if len(parts) == 2 && parts[1] == "variables" {
		s.handlers.HandleDeploymentVariables(w, r, deployment)
		return
	}

	if len(parts) >= 3 && parts[1] == "jobs" {
		job := parts[2]
		if len(parts) == 4 {
			job = parts[2] + "/" + parts[3]
		}
		s.handlers.HandleDeploymentJobs(w, r, deployment, job)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

// routeTasks routes task-related requests.
func (s *Server) routeTasks(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/tasks" {
		s.handlers.HandleTasks(w, r)
		return
	}

	parts := strings.Split(strings.TrimPrefix(path, "/tasks/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "task ID required")
		return
	}

	taskID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	if len(parts) == 1 {
		s.handlers.HandleTask(w, r, taskID)
		return
	}

	if len(parts) == 2 && parts[1] == "output" {
		s.handlers.HandleTaskOutput(w, r, taskID)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

// loggingMiddleware logs all requests.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		if s.config.Debug {
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
		}
	})
}

// authMiddleware validates Basic Auth.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/info" {
			next.ServeHTTP(w, r)
			return
		}

		if !s.handlers.CheckAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="BOSH Director"`)
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) generateTLSConfig() (*tls.Config, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Mock BOSH Director"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
