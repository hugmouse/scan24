package handler

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hugmouse/scan24/internal/parser"
	"golang.org/x/net/html/charset"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
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

var (
	tmplIndex    *template.Template
	tmplResult   *template.Template
	tmplProgress *template.Template
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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

func ResultHandler(w http.ResponseWriter, r *http.Request) {
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

	val, ok := globalMap[targetURL]
	if !ok {
		respondWithError(w, http.StatusBadRequest, tmplResult, "We don't have scan results for the following URL: "+baseURL.String())
		return
	}
	err = tmplResult.Execute(w, val)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing result template: %v", err), http.StatusInternalServerError)
		log.Printf("Error executing result template: %v", err)
	}
	return
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
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

	val, ok := globalMap[targetURL]
	if ok {
		err := tmplProgress.Execute(w, val)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error executing result template: %v", err), http.StatusInternalServerError)
			log.Printf("Error executing result template: %v", err)
		}
		return
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		respondWithError(w, http.StatusBadGateway, tmplResult, fmt.Sprintf("Failed to fetch URL: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondWithError(w, http.StatusBadGateway, tmplResult, fmt.Sprintf("URL that you provided failed to load, code: %d", resp.StatusCode))
		return
	}

	ct := resp.Header.Get("Content-Type")
	reader, err := charset.NewReader(resp.Body, ct)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, tmplResult, fmt.Sprintf("Failed to create charset reader: %v", err))
		return
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, tmplResult, fmt.Sprintf("Failed to read response body: %v", err))
		return
	}

	goqueryReader := bytes.NewReader(bodyBytes)

	doc, err := goquery.NewDocumentFromReader(goqueryReader)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, tmplResult, fmt.Sprintf("Failed to parse HTML: %v", err))
		return
	}

	// Start job, return status
	go doJob(doc, bodyBytes, baseURL, targetURL)
	globalMap[targetURL] = globalMapData{
		Page:     PageData{},
		URL:      targetURL,
		Progress: 0.0,
	}
	err = tmplProgress.Execute(w, globalMapData{
		Page:     PageData{},
		URL:      targetURL,
		Progress: 0.0,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing result template: %v", err), http.StatusInternalServerError)
		log.Printf("Error executing result template: %v", err)
	}
}

func doJob(doc *goquery.Document, bodyBytes []byte, baseURL *url.URL, targetURL string) {
	htmlVersion, err := parser.GetHTMLVersion(string(bodyBytes))
	if err != nil {
		log.Printf("Could not determine HTML version: %v", err)
		htmlVersion = &parser.DoctypeNode{Name: "Unknown"}
	}

	title := getTitle(doc)
	headings := make(map[string]int, 6)
	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		headings[tag] = doc.Find(tag).Length()
	}

	links := make([]parser.HyperLink, 0)

	var (
		jobCounter      int64
		jobDone         int64
		externalCounter int64
		externalAlive   int64
		internalCounter int64
		internalAlive   int64
		protocolCounter int64
	)

	var (
		wg     sync.WaitGroup
		linkMu sync.Mutex
	)
	// TODO: replace local ticker with global/some other mechanism of rate-limiting
	var rate = time.Tick(time.Second / time.Duration(1))

	// Essentially just run a goroutine for every <a> with a valid href value
	//
	// This also runs for href that = "#!" or "./" and etc since there might
	// be a custom HTTP server that renders something different on those URLs
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		attr, exists := s.Attr("href")
		if !exists {
			return
		}

		wg.Add(1)
		atomic.AddInt64(&jobCounter, 1)
		go func(attr string) {
			<-rate
			defer wg.Done()

			link := parser.Analyze(attr, baseURL)

			switch link.HrefType {
			case "external":
				atomic.AddInt64(&externalCounter, 1)
				if link.StatusCode == 200 {
					atomic.AddInt64(&externalAlive, 1)
				}
			case "internal":
				atomic.AddInt64(&internalCounter, 1)
				if link.StatusCode == 200 {
					atomic.AddInt64(&internalAlive, 1)
				}
			case "protocol":
				atomic.AddInt64(&protocolCounter, 1)
			}

			linkMu.Lock()
			links = append(links, link)
			linkMu.Unlock()
			atomic.AddInt64(&jobDone, 1)
			globalMap[targetURL] = globalMapData{
				Page:     PageData{},
				URL:      targetURL,
				Progress: float64(jobDone) / float64(jobCounter) * 100,
			}
		}(attr)
	})

	wg.Wait()

	haveLoginForm := parser.HasLoginForm(doc)

	globalMap[targetURL] = globalMapData{
		Page: PageData{
			URL:          targetURL,
			Title:        title,
			HTMLVersion:  htmlVersion.Name,
			Headings:     headings,
			HyperLinks:   links,
			HasLoginForm: haveLoginForm,
			LinkCounters: LinkCounters{
				Internal:      internalCounter,
				InternalAlive: internalAlive,
				External:      externalCounter,
				ExternalAlive: externalAlive,
				Protocol:      protocolCounter,
			},
		},
		URL:      targetURL,
		Progress: 100.0,
	}
}

var globalMap = make(map[string]globalMapData)

type globalMapData struct {
	Page     PageData
	URL      string
	Progress float64
	Error    string
}

func JobStatus(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		respondWithError(w, http.StatusBadRequest, tmplResult, "URL parameter is missing.")
		return
	}
	val, ok := globalMap[targetURL]
	if !ok {
		respondWithError(w, http.StatusBadRequest, tmplResult, "Job does not exist")
	}
	err := tmplProgress.Execute(w, val)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing result template: %v", err), http.StatusInternalServerError)
		log.Printf("Error executing result template: %v", err)
	}
}

// respondWithError is a helper function to streamline error handling
func respondWithError(w http.ResponseWriter, statusCode int, tmpl *template.Template, errMsg string) {
	w.WriteHeader(statusCode)
	data := PageData{Error: errMsg}
	_ = tmpl.Execute(w, data)
}

// getTitle attempts to get title from passed *goquery.Document
//
// Based on webkit source code HTML can contain multiple <title> tags,
// but we only care about the first one that we encounter
//
// Example:
//
// void Document::setTitleElement(const StringWithDirection& title, Element* titleElement)
//
//	{
//	    if (titleElement != m_titleElement) {
//	        if (m_titleElement || m_titleSetExplicitly)
//	            // Only allow the first title element to change the title -- others have no effect.
//	            return;
//	        m_titleElement = titleElement;
//	    }
//	    updateTitle(title);
//	}
func getTitle(doc *goquery.Document) (title string) {
	return doc.Find("title").First().Text()
}
