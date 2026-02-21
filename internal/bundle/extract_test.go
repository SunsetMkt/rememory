package bundle

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eljojo/rememory/internal/core"
)

func testShare() *core.Share {
	return core.NewShare(2, 1, 3, 2, "Alice", []byte("test-share-data"))
}

func TestExtractShareFromZip(t *testing.T) {
	share := testShare()
	readme := fmt.Sprintf("Some instructions\n\n%s\nMore text", share.Encode())

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "README.txt", Content: []byte(readme), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	got, err := ExtractShareFromZip(zipPath)
	if err != nil {
		t.Fatalf("ExtractShareFromZip: %v", err)
	}

	if got.Index != share.Index {
		t.Errorf("index: got %d, want %d", got.Index, share.Index)
	}
	if got.Total != share.Total {
		t.Errorf("total: got %d, want %d", got.Total, share.Total)
	}
	if got.Threshold != share.Threshold {
		t.Errorf("threshold: got %d, want %d", got.Threshold, share.Threshold)
	}
	if got.Holder != share.Holder {
		t.Errorf("holder: got %q, want %q", got.Holder, share.Holder)
	}
}

func TestExtractShareFromZip_LocalizedReadme(t *testing.T) {
	share := testShare()
	readme := share.Encode()

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "LEEME.txt", Content: []byte(readme), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	got, err := ExtractShareFromZip(zipPath)
	if err != nil {
		t.Fatalf("ExtractShareFromZip: %v", err)
	}

	if got.Index != share.Index {
		t.Errorf("index: got %d, want %d", got.Index, share.Index)
	}
}

func TestExtractShareFromZip_NoReadme(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "recover.html", Content: []byte("<html></html>"), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	_, err = ExtractShareFromZip(zipPath)
	if err == nil {
		t.Fatal("expected error for zip without README.txt")
	}
}

func TestExtractShareFromZip_NotAZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-zip.zip")
	if err := writeTestFile(path, []byte("not a zip")); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	_, err := ExtractShareFromZip(path)
	if err == nil {
		t.Fatal("expected error for non-zip file")
	}
}

func TestExtractManifestFromZip_DirectManifest(t *testing.T) {
	manifestData := []byte("encrypted-manifest-content")

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "MANIFEST.age", Content: manifestData, ModTime: time.Now()},
		{Name: "recover.html", Content: []byte("<html></html>"), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	got, err := ExtractManifestFromZip(zipPath)
	if err != nil {
		t.Fatalf("ExtractManifestFromZip: %v", err)
	}

	if string(got) != string(manifestData) {
		t.Errorf("manifest: got %q, want %q", got, manifestData)
	}
}

func TestExtractManifestFromZip_EmbeddedInHTML(t *testing.T) {
	manifestData := []byte("encrypted-manifest-content")
	b64 := base64.StdEncoding.EncodeToString(manifestData)

	personalization := map[string]interface{}{
		"manifestB64": b64,
		"share":       "test",
	}
	pJSON, _ := json.Marshal(personalization)
	htmlContent := fmt.Sprintf(`<html><script>window.PERSONALIZATION = %s;</script></html>`, pJSON)

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "recover.html", Content: []byte(htmlContent), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	got, err := ExtractManifestFromZip(zipPath)
	if err != nil {
		t.Fatalf("ExtractManifestFromZip: %v", err)
	}

	if string(got) != string(manifestData) {
		t.Errorf("manifest: got %q, want %q", got, manifestData)
	}
}

func TestExtractManifestFromZip_NoManifest(t *testing.T) {
	share := testShare()

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "README.txt", Content: []byte(share.Encode()), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	_, err = ExtractManifestFromZip(zipPath)
	if err == nil {
		t.Fatal("expected error for zip without manifest")
	}
}

func TestExtractManifestFromZip_HTMLWithoutEmbeddedManifest(t *testing.T) {
	// recover.html present but without embedded manifest (like --no-embed-manifest)
	htmlContent := `<html><script>window.PERSONALIZATION = {"share":"test","manifestB64":""};</script></html>`

	zipPath := filepath.Join(t.TempDir(), "bundle.zip")
	err := CreateZip(zipPath, []ZipFile{
		{Name: "recover.html", Content: []byte(htmlContent), ModTime: time.Now()},
	})
	if err != nil {
		t.Fatalf("creating zip: %v", err)
	}

	_, err = ExtractManifestFromZip(zipPath)
	if err == nil {
		t.Fatal("expected error for HTML without embedded manifest")
	}
}

func writeTestFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
