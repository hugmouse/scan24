package parser

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type HrefType string

var (
	ErrUnsupportedDoctype  = errors.New("unsupported DOCTYPE or quirky document (no <!DOCTYPE>)")
	ErrMalformedDoctype    = errors.New("malformed <!DOCTYPE> (no closing '>')")
	ErrUnterminatedQuote   = errors.New("unterminated quote in <!DOCTYPE>")
	ErrDoctypeNoNameToken  = errors.New("DOCTYPE has no name token, document considered quirky")
	ErrUnrecognizedDoctype = errors.New("unsupported or unrecognized DOCTYPE format, document considered quirky")
	ErrUnsupportedScheme   = errors.New("unsupported scheme")
)

const (
	Protocol HrefType = "protocol"
	External HrefType = "external"
	Internal HrefType = "internal"
)

func Classify(href string) HrefType {
	if isProtocol(href) {
		return Protocol
	}

	if isExternal(href) {
		return External
	}

	return Internal
}

// isProtocol checks if href is a protocol.
//
// Protocol is a link that has a scheme and NO host.
func isProtocol(href string) bool {
	u, err := url.Parse(href)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host == ""
}

// isExternal checks if href is an external link.
func isExternal(href string) bool {
	u, err := url.Parse(href)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

// HyperLink holds the analysis result.
type HyperLink struct {
	Raw        string
	Resolved   *url.URL
	HrefType   string
	StatusCode int
	Err        error
}

var (
	allowedSchemes = map[string]struct{}{
		"http":  {},
		"https": {},
	}
)

// Analyze inspects rawHref, resolves it against baseURL (if relative),
// then attempts a HEAD fetch while falling back to GET on 405 (Method Not Allowed).
//
// Returns following HTTP codes:
// -1 for unsupported schemes,
// 0 if any error occurred during fetching,
// or the actual HTTP status code otherwise.
func Analyze(rawHref string, baseURL *url.URL, httpClient *http.Client) HyperLink {
	hrefType := Classify(rawHref)

	// 1) parse & resolve
	resolved, err := resolveURL(rawHref, baseURL)
	if err != nil {
		return HyperLink{Raw: rawHref, HrefType: string(hrefType), Err: err}
	}

	// 2) check scheme
	if _, ok := allowedSchemes[resolved.Scheme]; !ok {
		return HyperLink{
			Raw:        rawHref,
			Resolved:   resolved,
			HrefType:   string(hrefType),
			Err:        ErrUnsupportedScheme,
			StatusCode: -1,
		}
	}

	// 3) and fetch it
	status, fetchErr := fetchStatus(resolved.String(), httpClient)
	if fetchErr != nil {
		log.Printf("Analyze: failed fetching %q: %v", resolved.String(), fetchErr)

		return HyperLink{
			Raw:        rawHref,
			Resolved:   resolved,
			HrefType:   string(hrefType),
			Err:        fetchErr,
			StatusCode: 0,
		}
	}

	return HyperLink{
		Raw:        rawHref,
		Resolved:   resolved,
		HrefType:   string(hrefType),
		StatusCode: status,
	}
}

// resolveURL parses rawHref and, if relative, resolves it against baseURL.
func resolveURL(rawHref string, baseURL *url.URL) (*url.URL, error) {
	_url, err := url.Parse(rawHref)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if _url.IsAbs() {
		return _url, nil
	}

	if baseURL == nil {
		return nil, fmt.Errorf("relative URL %q with nil base", rawHref)
	}

	return baseURL.ResolveReference(_url), nil
}

// fetchStatus does HEAD first; if it returns 405 Method Not Allowed,
// it retries with GET.
func fetchStatus(url string, httpClient *http.Client) (int, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Scan24 (+https://github.com/hugmouse/scan24)")
	// HEAD
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to HEAD the url '%s': %w", url, err)
	}
	defer resp.Body.Close()

	// retry with GET instead
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusBadRequest {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err
		}
		resp2, err2 := httpClient.Do(req)
		if err2 != nil {
			return 0, fmt.Errorf("failed to GET the url '%s': %w", url, err)
		}

		defer resp2.Body.Close()

		return resp2.StatusCode, nil
	}

	return resp.StatusCode, nil
}
