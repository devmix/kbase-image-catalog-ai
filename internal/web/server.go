package web

import (
	"context"
	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
	"kbase-catalog/internal/web/api"
	"kbase-catalog/web"
	"log"
	"net/http"
	"strconv"
)

// Server represents the web server
type Server struct {
	config     *config.Config
	port       int
	httpServer *http.Server
	apiHandler *api.APIHandler
}

// NewServer creates a new web server instance
func NewServer(cfg *config.Config, catalogProcessor *processor.CatalogProcessor, port int, archivePath string) *Server {
	apiHandler, err := api.NewAPIHandler(cfg, catalogProcessor, archivePath)
	if err != nil {
		log.Printf("Failed to create API handler: %v", err)
	}

	return &Server{
		config:     cfg,
		port:       port,
		apiHandler: apiHandler,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static files handler for images
	mux.HandleFunc("/archive/", s.apiHandler.HandleArchiveFiles)

	// Static files handler for static assets
	mux.HandleFunc("/static/", web.HandleEmbeddedFile)

	// Web interface handlers
	mux.HandleFunc("/", s.apiHandler.HandleIndex)
	mux.HandleFunc("/api/catalog", s.apiHandler.HandleApiCatalog)
	mux.HandleFunc("/api/search", s.apiHandler.HandleApiSearch)
	mux.HandleFunc("/api/reindex", s.apiHandler.HandleReindex)
	mux.HandleFunc("/api/catalog-search", s.apiHandler.HandleApiCatalogSearch)
	mux.HandleFunc("/catalog/", s.apiHandler.HandleCatalogDetail)

	// Apply middleware
	var handler http.Handler = mux
	handler = api.LoggingMiddleware(handler)
	handler = api.RecoveryMiddleware(handler)
	handler = api.CORSMiddleware(handler)

	s.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(s.port),
		Handler: handler,
	}

	log.Printf("Starting web server on http://localhost:%d\n", s.port)

	if err := s.apiHandler.Start(); err != nil {
		return err
	}

	// Start the server in a goroutine so we can handle shutdown signals
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the web server
func (s *Server) Stop(ctx context.Context) error {
	s.apiHandler.Stop()
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
