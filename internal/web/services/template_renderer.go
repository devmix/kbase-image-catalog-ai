package services

import (
	"fmt"
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

// RenderCatalogList renders the HTML for catalog lists
func (tr *TemplateRenderer) RenderCatalogList(catalogs []map[string]interface{}) template.HTML {
	var html strings.Builder
	html.WriteString("<div class=\"catalog-grid\">\n")

	if len(catalogs) == 0 {
		html.WriteString("<p>No catalogs found.</p>\n")
	} else {
		for _, catalog := range catalogs {
			name, _ := catalog["name"].(string)
			imageCount, _ := catalog["imageCount"].(int)
			lastUpdate, _ := catalog["lastUpdate"].(string)

			// Format the last update date nicely if available
			formattedDate := ""
			if lastUpdate != "" {
				if t, err := time.Parse(time.RFC3339, lastUpdate); err == nil {
					formattedDate = t.Format("2006-01-02")
				} else {
					formattedDate = lastUpdate // fallback if parsing fails
				}
			}

			html.WriteString(fmt.Sprintf(`<div class="catalog-card"><a href="/catalog/%s"><h3>%s</h3><p>Images: %d</p><p>Last update: %s</p></a></div>`, name, name, imageCount, formattedDate))
		}
	}

	html.WriteString("</div>")
	return template.HTML(html.String())
}

// RenderCatalogNavigation renders navigation links for catalogs
func (tr *TemplateRenderer) RenderCatalogNavigation(catalogs []map[string]interface{}, current string) template.HTML {
	var html strings.Builder
	html.WriteString("<span>Catalogs: </span>")

	for _, catalog := range catalogs {
		name, _ := catalog["name"].(string)
		if name == current {
			html.WriteString(fmt.Sprintf(`<strong>%s</strong>`, name))
		} else {
			html.WriteString(fmt.Sprintf(`<a href="/catalog/%s">%s</a>`, name, name))
		}
		html.WriteString(" | ")
	}

	return template.HTML(html.String())
}

// RenderCatalogImages renders HTML for catalog images
func (tr *TemplateRenderer) RenderCatalogImages(catalogImages []map[string]interface{}, catalogName string) template.HTML {
	var html strings.Builder
	html.WriteString("<div class=\"image-grid\">\n")

	for _, imageData := range catalogImages {
		if filename, ok := imageData["filename"].(string); ok {
			shortName := filename
			description := ""

			if sn, ok := imageData["short_name"].(string); ok && sn != "" {
				shortName = sn
			}

			if desc, ok := imageData["description"].(string); ok && desc != "" {
				description = desc
			}

			html.WriteString(fmt.Sprintf(`
<div class="image-card">
    <img src="/archive/%s/%s" alt="%s" style="max-width: 100%%; height: auto;" />
    <div class="image-info">
        <div class="image-title">%s</div>
        <div class="image-description">%s</div>
    </div>
</div>`,
				catalogName,
				filename,
				shortName,
				shortName,
				description))
		}
	}

	html.WriteString("</div>")
	return template.HTML(html.String())
}
