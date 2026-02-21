package bundle

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/eljojo/rememory/internal/core"
	"github.com/eljojo/rememory/internal/html"
	"github.com/eljojo/rememory/internal/translations"
)

// ExtractShareFromZip opens a bundle ZIP and parses the share from the
// README.txt inside it.
func ExtractShareFromZip(zipPath string) (*core.Share, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !translations.IsReadmeFile(f.Name, ".txt") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s in zip: %w", f.Name, err)
		}

		data, err := io.ReadAll(rc)
		if closeErr := rc.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s from zip: %w", f.Name, err)
		}

		share, err := core.ParseShare(data)
		if err != nil {
			return nil, fmt.Errorf("parsing share from %s: %w", f.Name, err)
		}

		return share, nil
	}

	return nil, fmt.Errorf("no README file (.txt) found in zip")
}

// ExtractManifestFromZip opens a bundle ZIP and extracts the encrypted
// manifest data. It looks for MANIFEST.age first. If not found, it falls
// back to extracting the manifest embedded in recover.html.
func ExtractManifestFromZip(zipPath string) ([]byte, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	var recoverData []byte

	for _, f := range r.File {
		switch {
		case f.Name == "MANIFEST.age":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening MANIFEST.age in zip: %w", err)
			}

			data, err := io.ReadAll(rc)
			if closeErr := rc.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return nil, fmt.Errorf("reading MANIFEST.age from zip: %w", err)
			}
			return data, nil

		case f.Name == "recover.html":
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening recover.html in zip: %w", err)
			}

			recoverData, err = io.ReadAll(rc)
			if closeErr := rc.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return nil, fmt.Errorf("reading recover.html from zip: %w", err)
			}
		}
	}

	// Fall back to extracting manifest from recover.html personalization data
	if len(recoverData) > 0 {
		manifest, err := html.ExtractManifestFromHTML(recoverData)
		if err != nil {
			return nil, fmt.Errorf("extracting manifest from recover.html in zip: %w", err)
		}
		return manifest, nil
	}

	return nil, fmt.Errorf("no MANIFEST.age or recover.html found in zip")
}
