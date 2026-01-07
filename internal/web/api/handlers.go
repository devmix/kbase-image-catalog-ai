package api

import (
	"encoding/json"
	"kbase-catalog/internal/errors"
	"kbase-catalog/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
	"kbase-catalog/internal/web/queue"
	"kbase-catalog/internal/web/services"
	"kbase-catalog/internal/web/watch"
)

// APIHandler represents the API handlers
type APIHandler struct {
	config           *config.Config
	processor        *processor.CatalogProcessor
	catalogService   *services.CatalogService
	templateRenderer *services.TemplateRenderer
	taskQueue        *queue.TaskQueue
	watcher          *watch.CatalogWatcher
	archivePath      string
}

// NewAPIHandler creates a new API handler instance
func NewAPIHandler(cfg *config.Config, catalogProcessor *processor.CatalogProcessor, archivePath string) (*APIHandler, error) {
	taskQueue := queue.NewTaskQueue(cfg, catalogProcessor, archivePath)
	watcher, err := watch.NewCatalogWatcher(taskQueue, archivePath)
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
	}

	catalogService := &services.CatalogService{Config: cfg, Processor: catalogProcessor, ArchiveDir: archivePath}

	return &APIHandler{
		config:           cfg,
		processor:        catalogProcessor,
		catalogService:   catalogService,
		templateRenderer: services.NewTemplateRenderer(catalogService),
		taskQueue:        taskQueue,
		watcher:          watcher,
		archivePath:      archivePath,
	}, nil
}

// HandleIndex serves the main index page
func (h *APIHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get sort parameters from query string for index page catalogs
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	catalogs, err := h.catalogService.GetCatalogs(r.Context())
	if err != nil {
		log.Printf("Error getting catalogs for index: %v", err)
		http.Error(w, "Failed to load catalog list", http.StatusInternalServerError)
		return
	}

	catalogs = SortCatalogs(catalogs, sortBy, sortOrder)

	err = h.templateRenderer.RenderTemplate(w, r, "templates/index.html", "templates/catalog-list-fragment.html", map[string]interface{}{
		"CatalogList": h.templateRenderer.RenderCatalogList(catalogs),
	})
	if err != nil {
		return // Error already handled by RenderTemplate
	}
}

// HandleApiCatalog returns list of all catalogs with extra information as JSON
func (h *APIHandler) HandleApiCatalog(w http.ResponseWriter, r *http.Request) {
	// Get sort parameters from query string
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	catalogs, err := h.catalogService.GetCatalogs(r.Context())
	if err != nil {
		log.Printf("Error getting catalogs: %v", err)
		http.Error(w, "Failed to retrieve catalogs", http.StatusInternalServerError)
		return
	}

	catalogs = SortCatalogs(catalogs, sortBy, sortOrder)

	jsonData, err := json.Marshal(catalogs)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// HandleApiSearch returns search results as HTML or JSON
func (h *APIHandler) HandleApiSearch(w http.ResponseWriter, r *http.Request) {
	// Try to get query parameter first
	query := r.URL.Query().Get("q")

	// Also check form values in case HTMX sends it differently
	if query == "" {
		r.ParseForm()
		query = r.FormValue("q")
	}

	log.Printf("Search query received: '%s'", query)

	// Get sort parameters from query string for search results
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	catalogs, err := h.catalogService.SearchCatalogs(r.Context(), query)
	if err != nil {
		log.Printf("Error during search: %v", err)
		http.Error(w, "Failed to perform search", http.StatusInternalServerError)
		return
	}

	catalogs = SortCatalogs(catalogs, sortBy, sortOrder)

	err = h.templateRenderer.RenderTemplate(w, r, "templates/search-result.html", "templates/catalog-list-fragment.html", map[string]interface{}{
		"CatalogList": h.templateRenderer.RenderCatalogList(catalogs),
	})
	if err != nil {
		return // Error already handled by RenderTemplate
	}
}

// HandleApiCatalogSearch handles searching for images within a specific catalog
func (h *APIHandler) HandleApiCatalogSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get catalog name and search query from URL parameters
	catalogName := r.URL.Query().Get("catalog")
	query := r.URL.Query().Get("q")

	log.Printf("Catalog search query received: catalog='%s', query='%s'", catalogName, query)

	if catalogName == "" {
		http.Error(w, "Missing 'catalog' parameter", http.StatusBadRequest)
		return
	}

	// Get sort parameters from query string for search results
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	// Search within the specific catalog
	indexData, err := h.catalogService.SearchCatalogImages(r.Context(), catalogName, query)
	if err != nil {
		log.Printf("Error during catalog search: %v", err)
		http.Error(w, "Failed to perform catalog search", http.StatusInternalServerError)
		return
	}

	sortedIndexData := SortCatalogImages(indexData, sortBy, sortOrder)

	// For non-HTMX requests, return JSON response
	isHTMX := r.Header.Get("HX-Request") == "true"
	if !isHTMX {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(indexData)
		return
	}

	// For HTMX requests, render the fragment
	err = h.templateRenderer.RenderTemplate(w, r, "", "templates/catalog-images-fragment.html", map[string]interface{}{
		"CatalogImages": h.templateRenderer.RenderCatalogImages(sortedIndexData, catalogName),
	})
	if err != nil {
		return // Error already handled by RenderTemplate
	}
}

// HandleCatalogDetail serves individual catalog detail pages
func (h *APIHandler) HandleCatalogDetail(w http.ResponseWriter, r *http.Request) {
	catalogName := strings.TrimPrefix(r.URL.Path, "/catalog/")

	if catalogName == "" {
		http.NotFound(w, r)
		return
	}

	// Get sort parameters from query string
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	// Get the index.json for this catalog
	indexData, err := h.catalogService.GetCatalogImages(r.Context(), catalogName)
	if err != nil {
		log.Printf("Error getting catalog images: %v", err)
		http.NotFound(w, r)
		return
	}

	sortedIndexData := SortCatalogImages(indexData, sortBy, sortOrder)

	err = h.templateRenderer.RenderTemplate(w, r, "templates/catalog-detail.html", "templates/catalog-images-fragment.html", map[string]interface{}{
		"CatalogName":   catalogName,
		"CatalogImages": h.templateRenderer.RenderCatalogImages(sortedIndexData, catalogName),
	})
	if err != nil {
		return // Error already handled by RenderTemplate
	}
}

// HandleReindex handles manual reindex requests
func (h *APIHandler) HandleReindex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form data: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	catalogName := r.FormValue("catalog")

	// If catalogName is empty, reindex all catalogs
	if catalogName == "" {
		// Get all catalogs
		catalogs, err := h.catalogService.GetCatalogs(r.Context())
		if err != nil {
			log.Printf("Error getting catalogs for reindex: %v", err)
			http.Error(w, "Failed to get catalog list", http.StatusInternalServerError)
			return
		}

		// Add tasks for each catalog to the queue
		for _, catalog := range catalogs {
			if name, ok := catalog["name"].(string); ok && name != "" {
				if err := h.taskQueue.AddTask(name, "manual"); err != nil {
					log.Printf("Failed to add reindex task for catalog %s: %v", name, err)
				} else {
					log.Printf("Reindex task queued for catalog: %s", name)
				}
			}
		}

		// For HTMX requests, return a simple HTML message instead of JSON
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<span class="alert alert-success">Reindex tasks queued for all catalogs</span>`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "Reindex tasks queued for all catalogs",
			})
		}
		return
	}

	// Add the reindex task to the queue for specific catalog
	if err := h.taskQueue.AddTask(catalogName, "manual"); err != nil {
		log.Printf("Failed to add reindex task: %v", err)
		http.Error(w, "Failed to queue reindex task", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return a simple HTML message instead of JSON
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<span class="alert alert-success">Reindex task queued for catalog: ` + catalogName + `</span>`))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Reindex task queued for catalog: " + catalogName,
		})
	}
}

// HandleArchiveFiles serves static files from the archive directory
func (h *APIHandler) HandleArchiveFiles(w http.ResponseWriter, r *http.Request) {
	// Serve files from archive directory
	path := strings.TrimPrefix(r.URL.Path, "/archive/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	// Construct the full file path using configured archive directory
	fullPath := h.archivePath + "/" + path

	// Check if file exists
	if !utils.IsFileExists(fullPath) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// HandleStaticFiles serves static files from the web/static directory
func (h *APIHandler) HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Serve files from web/static directory
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	// Construct the full file path
	fullPath := "web/static/" + path

	// Check if file exists
	if !utils.IsFileExists(fullPath) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

func (h *APIHandler) Start() *errors.WebServerError {
	// Start the task queue
	if err := h.taskQueue.Start(); err != nil {
		log.Printf("Failed to start task queue: %v", err)
		return &errors.WebServerError{
			BaseError: errors.BaseError{
				Code:      "FAIL_TO_START_TASKS_QUEUE",
				Message:   "Failed to start task queue",
				Timestamp: time.Now(),
				Details:   err.Error(),
			},
		}
	} else {
		log.Printf("Task queue started successfully")
	}

	// Start the file watcher
	if h.watcher != nil {
		if err := h.watcher.Start(); err != nil {
			log.Printf("Failed to start file watcher: %v", err)
			return &errors.WebServerError{
				BaseError: errors.BaseError{
					Code:      "FAIL_TO_START_CATALOG_WATCHER",
					Message:   "Failed to start catalog watcher",
					Timestamp: time.Now(),
					Details:   err.Error(),
				},
			}
		} else {
			log.Printf("File watcher started successfully")
		}
	} else {
		log.Printf("No file watcher created - check configuration")
	}

	return nil
}

func (h *APIHandler) Stop() {
	// Stop the watcher first
	if h.watcher != nil {
		h.watcher.Stop()
	}

	// Stop the task queue
	if h.taskQueue != nil {
		h.taskQueue.Stop()
	}
}
