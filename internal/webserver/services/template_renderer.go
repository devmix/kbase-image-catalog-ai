package services

import (
	"html/template"
	"kbase-catalog/web"
	"log"
	"net/http"
	"strings"
	"time"
)

// TemplateRenderer handles template rendering operations
type TemplateRenderer struct {
	catalogService *CatalogService
}

// NewTemplateRenderer creates a new template renderer instance
func NewTemplateRenderer(catalogService *CatalogService) *TemplateRenderer {
	return &TemplateRenderer{
		catalogService: catalogService,
	}
}

// RenderTemplate handles rendering of templates with HTMX support
func (tr *TemplateRenderer) RenderTemplate(w http.ResponseWriter, r *http.Request, fullTemplatePath, fragmentTemplatePath string, data map[string]interface{}) error {
	isHTMX := r.Header.Get("HX-Request") == "true"

	if isHTMX && fragmentTemplatePath != "" {
		// For HTMX requests, only render the fragment
		tmpl, err := template.ParseFS(web.FS, fragmentTemplatePath)
		if err != nil {
			log.Printf("Failed to load fragment template %s: %v", fragmentTemplatePath, err)
			http.Error(w, "Failed to load template", http.StatusInternalServerError)
			return err
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing fragment template %s: %v", fragmentTemplatePath, err)
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			return err
		}
	} else {
		// For regular requests, render the full template
		tmpl, err := template.ParseFS(web.FS, fullTemplatePath)
		if err != nil {
			log.Printf("Failed to load template %s: %v", fullTemplatePath, err)
			http.Error(w, "Failed to load template", http.StatusInternalServerError)
			return err
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template %s: %v", fullTemplatePath, err)
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			return err
		}
	}

	return nil
}

// RenderCatalogList renders the HTML for catalog lists using a template
func (tr *TemplateRenderer) RenderCatalogList(catalogs []map[string]interface{}) template.HTML {
	// Format the data as needed by templates
	formattedCatalogs := make([]map[string]interface{}, len(catalogs))
	for i, catalog := range catalogs {
		data := map[string]interface{}{}

		// Copy all fields from original catalog
		for k, v := range catalog {
			data[k] = v
		}

		// Format the last update date nicely if available
		if lastUpdate, ok := catalog["lastUpdate"].(string); ok && lastUpdate != "" {
			if t, err := time.Parse(time.RFC3339, lastUpdate); err == nil {
				data["lastUpdate"] = t.Format("2006-01-02")
			} else {
				data["lastUpdate"] = lastUpdate // fallback if parsing fails
			}
		}

		formattedCatalogs[i] = data
	}

	data := map[string]interface{}{
		"CatalogList": formattedCatalogs,
	}

	tmpl, err := template.ParseFS(web.FS, "templates/catalog-list-template.html")
	if err != nil {
		log.Printf("Failed to load catalog list template: %v", err)
		return ""
	}

	var html strings.Builder
	err = tmpl.Execute(&html, data)
	if err != nil {
		log.Printf("Error executing catalog list template: %v", err)
		return ""
	}

	return template.HTML(html.String())
}

// RenderCatalogNavigation renders navigation links for catalogs using a template
func (tr *TemplateRenderer) RenderCatalogNavigation(catalogs []map[string]interface{}, current string) template.HTML {
	data := map[string]interface{}{
		"CatalogNavigation": catalogs,
		"CurrentCatalog":    current,
	}

	tmpl, err := template.ParseFS(web.FS, "templates/catalog-navigation-template.html")
	if err != nil {
		log.Printf("Failed to load catalog navigation template: %v", err)
		return ""
	}

	var html strings.Builder
	err = tmpl.Execute(&html, data)
	if err != nil {
		log.Printf("Error executing catalog navigation template: %v", err)
		return ""
	}

	return template.HTML(html.String())
}

// RenderCatalogImages renders HTML for catalog images using a template
func (tr *TemplateRenderer) RenderCatalogImages(catalogImages []map[string]interface{}, catalogName string) template.HTML {
	// Format the data as needed by templates
	formattedImages := make([]map[string]interface{}, len(catalogImages))
	for i, imageData := range catalogImages {
		data := map[string]interface{}{}

		if filename, ok := imageData["filename"].(string); ok {
			shortName := filename
			description := ""

			if sn, ok := imageData["short_name"].(string); ok && sn != "" {
				shortName = sn
			}

			if desc, ok := imageData["description"].(string); ok && desc != "" {
				description = desc
			}

			data["filename"] = filename
			data["title"] = shortName
			data["description"] = description
		}
		formattedImages[i] = data
	}

	data := map[string]interface{}{
		"catalog": catalogName,
		"images":  formattedImages,
	}

	tmpl, err := template.ParseFS(web.FS, "templates/catalog-images-template.html")
	if err != nil {
		log.Printf("Failed to load catalog images template: %v", err)
		return ""
	}

	var html strings.Builder
	err = tmpl.Execute(&html, data)
	if err != nil {
		log.Printf("Error executing catalog images template: %v", err)
		return ""
	}

	return template.HTML(html.String())
}
