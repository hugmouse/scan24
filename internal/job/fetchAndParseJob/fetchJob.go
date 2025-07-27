package fetchAndParseJob

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hugmouse/scan24/internal/handler"
	"github.com/hugmouse/scan24/internal/job"
	"github.com/hugmouse/scan24/internal/parser"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type FetchJob struct {
	ID        string
	URL       string
	Client    *http.Client
	RateLimit int
	Progress  float64
}

func (fj *FetchJob) GetID() string {
	return fj.ID
}

func (fj *FetchJob) GetType() string {
	return "FetchJob"
}

func (fj *FetchJob) GetProgress() float64 {
	return fj.Progress
}

type PageAnalysisData struct {
	URL          string
	Title        string
	HTMLVersion  string
	Headings     map[string]int
	HyperLinks   []parser.HyperLink
	HasLoginForm bool
	LinkCounters struct {
		Internal      int64
		InternalAlive int64
		External      int64
		ExternalAlive int64
		Protocol      int64
	}
}

func (fj *FetchJob) Execute() job.Result {
	log.Printf("Executing FetchJob %s: Fetching %s\n", fj.ID, fj.URL)

	resp, err := fj.Client.Get(fj.URL)
	if err != nil {
		return job.Result{
			JobID:   fj.ID,
			JobType: fj.GetType(),
			Error:   fmt.Errorf("fetch error: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return job.Result{
			JobID:   fj.ID,
			JobType: fj.GetType(),
			Error:   fmt.Errorf("fetch failed with status: %s", resp.Status),
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return job.Result{
			JobID:   fj.ID,
			JobType: fj.GetType(),
			Error:   fmt.Errorf("body read error: %w", err),
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return job.Result{
			JobID:   fj.ID,
			JobType: fj.GetType(),
			Error:   fmt.Errorf("document parse error: %w", err),
		}
	}

	baseURL := resp.Request.URL

	htmlVersion, err := parser.GetHTMLVersion(string(bodyBytes))
	if err != nil {
		log.Printf("Could not determine HTML version for %s: %v", fj.URL, err)
		htmlVersion = &parser.DoctypeNode{Name: "Unknown"}
	}

	title := getTitle(doc)

	headings := make(map[string]int, 6)
	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		headings[tag] = doc.Find(tag).Length()
	}

	links := make([]parser.HyperLink, 0)
	var (
		internalCounter, internalAlive int64
		externalCounter, externalAlive int64
		protocolCounter                int64
	)

	var wg sync.WaitGroup
	var linkMu sync.Mutex
	rate := time.Tick(time.Second / time.Duration(fj.RateLimit))

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		attr, _ := s.Attr("href") // Existence already checked by selector
		wg.Add(1)

		go func(attr string) {
			defer wg.Done()
			<-rate

			link := parser.Analyze(attr, baseURL, fj.Client)

			switch link.HrefType {
			case "external":
				atomic.AddInt64(&externalCounter, 1)
				if link.StatusCode == http.StatusOK {
					atomic.AddInt64(&externalAlive, 1)
				}
			case "internal":
				atomic.AddInt64(&internalCounter, 1)
				if link.StatusCode == http.StatusOK {
					atomic.AddInt64(&internalAlive, 1)
				}
			case "protocol":
				atomic.AddInt64(&protocolCounter, 1)
			}

			linkMu.Lock()
			links = append(links, link)
			linkMu.Unlock()
		}(attr)
	})

	wg.Wait()

	haveLoginForm := parser.HasLoginForm(doc)

	resultData := PageAnalysisData{
		URL:          fj.URL,
		Title:        title,
		HTMLVersion:  htmlVersion.Name,
		Headings:     headings,
		HyperLinks:   links,
		HasLoginForm: haveLoginForm,
		LinkCounters: handler.LinkCounters{
			Internal:      internalCounter,
			InternalAlive: internalAlive,
			External:      externalCounter,
			ExternalAlive: externalAlive,
			Protocol:      protocolCounter,
		},
	}

	return job.Result{
		JobID:   fj.ID,
		JobType: fj.GetType(),
		Data:    resultData,
		Error:   nil,
	}
}

func getTitle(doc *goquery.Document) (title string) {
	return doc.Find("title").First().Text()
}
