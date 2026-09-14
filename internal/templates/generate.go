package templates

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateReports renders all page templates and writes the resulting HTML files to outputDir.
func (r *PageRenderer) GenerateReports(outputDir string, data any) error {
	pages, err := ListPages()
	if err != nil {
		return fmt.Errorf("listing pages: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for _, page := range pages {
		if err := r.generatePage(outputDir, page, data); err != nil {
			return err
		}
	}
	return nil
}

// generatePage renders a single page and writes it to outputDir.
func (r *PageRenderer) generatePage(outputDir, pageName string, data any) error {
	path := filepath.Join(outputDir, pageName+".html")

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}
	defer f.Close()

	if err := r.Render(f, pageName, data); err != nil {
		return fmt.Errorf("rendering page %s: %w", pageName, err)
	}
	return nil
}