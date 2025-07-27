package handler

import (
	"fmt"
	"github.com/hugmouse/scan24/internal/job"
	"github.com/hugmouse/scan24/internal/job/fetchAndParseJob"
	"github.com/hugmouse/scan24/internal/parser"
	"github.com/hugmouse/scan24/internal/workerPool"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type LinkCounters struct {
	Internal      int64
	InternalAlive int64
	External      int64
	ExternalAlive int64
	Protocol      int64
}

type PageData struct {
	URL             string
	Title           string
	HTMLVersion     string
	Headings        map[string]int
	LinkCounters    LinkCounters
	HyperLinks      []parser.HyperLink
	HasLoginForm    bool
	SiteInformation string
	Error           string
}

type globalMapData struct {
	Page     PageData
	URL      string
	Progress float64
	Error    string
}

var (
	tmplIndex    *template.Template
	tmplResult   *template.Template
	tmplProgress *template.Template
	globalMap    = make(map[string]globalMapData)
	globalLock   = sync.Mutex{}
)

type Handler struct {
	Client     *http.Client
	RateLimit  int
	WorkerPool *workerPool.WorkerPool
	JobResults map[string]job.Result
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	err := tmplIndex.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		log.Printf("Error executing index template: %v", err)
	}
}

func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, tmplResult, "Method not allowed. Use GET.")

		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		respondWithError(w, http.StatusBadRequest, tmplResult, "URL parameter is missing.")

		return
	}

	baseURL, err := url.ParseRequestURI(targetURL)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, tmplResult, fmt.Sprintf("Invalid URL provided: %v", err))

		return
	}

	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		respondWithError(w, http.StatusBadRequest, tmplResult, "Only HTTP/HTTPS links are allowed. For example: https://mysh.dev")

		return
	}

	// Do the job, return 202

	h.WorkerPool.SubmitJob(&fetchAndParseJob.FetchJob{
		ID:        baseURL.String(),
		URL:       baseURL.String(),
		Client:    h.Client,
		RateLimit: h.RateLimit,
	})

	w.WriteHeader(http.StatusAccepted)
	return
}

// respondWithError is a helper function to streamline error handling.
func respondWithError(w http.ResponseWriter, statusCode int, tmpl *template.Template, errMsg string) {
	w.WriteHeader(statusCode)

	data := PageData{Error: errMsg}
	_ = tmpl.Execute(w, data)
}
