package project

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/manifest-readme.md
var manifestReadmeTemplate string

// TemplateData contains data for rendering templates.
type TemplateData struct {
	ProjectName string
	Friends     []Friend
	Threshold   int
}

// WriteManifestReadme creates the README.md file in the manifest directory.
func WriteManifestReadme(manifestDir string, data TemplateData) error {
	content, err := RenderManifestReadme(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(manifestDir, "README.md"), content, 0644)
}

// TemplateDataFromProject builds TemplateData from a Project.
func TemplateDataFromProject(p *Project) TemplateData {
	return TemplateData{
		ProjectName: p.Name,
		Friends:     p.Friends,
		Threshold:   p.Threshold,
	}
}

// RenderManifestReadme renders the manifest README template to bytes.
// Used at seal time to detect an untouched template.
func RenderManifestReadme(data TemplateData) ([]byte, error) {
	tmpl, err := template.New("readme").Parse(manifestReadmeTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return []byte(buf.String()), nil
}

// FriendNames returns a comma-separated list of friend names.
func FriendNames(friends []Friend) string {
	names := make([]string, len(friends))
	for i, f := range friends {
		names[i] = f.Name
	}
	return strings.Join(names, ", ")
}
