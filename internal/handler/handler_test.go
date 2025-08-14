package handler

import (
	"fmt"
	"github.com/hugmouse/scan24/internal/cache"
	"github.com/hugmouse/scan24/internal/ratelimiter"
	"golang.org/x/time/rate"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeHandler_Cached(t *testing.T) {
	initTemplates(t)

	jobCache := cache.New[string, GlobalMapData](time.Minute * 1)

	h := &Handler{
		Cache: jobCache,
	}

	// Mock completed job
	testURL := "https://example.com"
	jobCache.Set(testURL, GlobalMapData{
		Page: PageData{
			Title: "Test Page",
		},
		URL:      testURL,
		Progress: 100.0,
	})

	req := httptest.NewRequest("GET", "/analyze?url="+testURL, nil)
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that the result template was executed
	if !strings.Contains(rr.Body.String(), "Test Page") {
		t.Errorf("handler returned unexpected body: got %v",
			rr.Body.String())
	}
}

func TestAnalyzeHandler_WithMockServer(t *testing.T) {
	initTemplates(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, `<!DOCTYPE html><html><head><title>Test</title></head><body><a href="/page2">Link</a></body></html>`)
	}))
	defer server.Close()

	jobCache := cache.New[string, GlobalMapData](time.Minute * 1)
	limiter := ratelimiter.NewDomainRateLimiter(rate.Limit(10), 1)
	h := &Handler{
		Client:      server.Client(),
		Cache:       jobCache,
		RateLimiter: limiter,
	}

	req := httptest.NewRequest("GET", "/analyze?url="+server.URL, nil)
	rr := httptest.NewRecorder()

	h.AnalyzeHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "progress-bar") {
		t.Errorf("handler returned unexpected body, expected progress bar: got %v",
			rr.Body.String())
	}

	// Check that the item was added to the cache
	_, ok := jobCache.Get(server.URL)
	if !ok {
		t.Errorf("expected item to be added to cache for url: %s", server.URL)
	}
}

func initTemplates(t *testing.T) {
	var err error
	tmplResult, err = template.New("result.gohtml").Parse(`Title: {{.Page.Title}}`)
	if err != nil {
		t.Fatalf("Failed to parse result.gohtml for test: %v", err)
	}
	tmplProgress, err = template.New("progress.gohtml").Parse(`<div class="progress-bar" style="width:{{.Progress}}%"></div>`)
	if err != nil {
		t.Fatalf("Failed to parse progress.gohtml for test: %v", err)
	}
}
