package serve

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eljojo/rememory/internal/core"
)

// Config holds the configuration for the server.
type Config struct {
	Host            string
	Port            string
	DataDir         string
	MaxManifestSize int  // Maximum MANIFEST.age size in bytes
	NoTlock         bool // Omit time-lock support
	Version         string
}

// Server implements http.Handler for the self-hosted ReMemory web app.
type Server struct {
	store           *Store
	maxManifestSize int
	noTlock         bool
	version         string
	githubURL       string
	mux             *http.ServeMux
}

// New creates a new Server from the given config.
func New(cfg Config) (*Server, error) {
	store, err := NewStore(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("initializing store: %w", err)
	}

	var githubURL string
	if strings.HasPrefix(cfg.Version, "v") {
		githubURL = fmt.Sprintf("%s/releases/tag/%s", core.GitHubRepo, cfg.Version)
	} else {
		githubURL = core.GitHubRepo + "/releases/latest"
	}

	s := &Server{
		store:           store,
		maxManifestSize: cfg.MaxManifestSize,
		noTlock:         cfg.NoTlock,
		version:         cfg.Version,
		githubURL:       githubURL,
		mux:             http.NewServeMux(),
	}

	s.routes()
	return s, nil
}

// routes registers all HTTP routes.
func (s *Server) routes() {
	// Pages
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /create", s.handleCreate)
	s.mux.HandleFunc("GET /recover", s.handleRecover)
	s.mux.HandleFunc("GET /about", s.handleAbout)
	s.mux.HandleFunc("GET /docs", s.handleDocs)

	// Redirect .html paths to clean routes (docs content links to these)
	for _, r := range [][2]string{
		{"/index.html", "/about"},
		{"/maker.html", "/create"},
		{"/recover.html", "/recover"},
		{"/docs.html", "/docs"},
	} {
		from, to := r[0], r[1]
		s.mux.HandleFunc("GET "+from, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, to, http.StatusMovedPermanently)
		})
	}

	// API
	s.mux.HandleFunc("GET /api/status", s.handleAPIStatus)
	s.mux.HandleFunc("POST /api/setup", s.handleAPISetup)
	s.mux.HandleFunc("GET /api/bundles", s.handleAPIListBundles)
	s.mux.HandleFunc("POST /api/bundle", s.handleAPISaveBundle)
	s.mux.HandleFunc("DELETE /api/bundle", s.handleAPIDeleteBundle)
	s.mux.HandleFunc("GET /api/bundle/manifest", s.handleAPIManifest)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// Store returns the server's store (for testing).
func (s *Server) Store() *Store {
	return s.store
}
