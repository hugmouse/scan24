package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	CommonAutocompleteValues = []string{"username", "current-password", "new-password", "one-time-code"}
	CommonLoginInputValues   = []string{"email", "text"}
)

// HasLoginForm checks if a document contains a login form
//
// It checks for any elements that have:
//
// - autocomplete="username"
// - autocomplete="current-password"
// - autocomplete="new-password"
// - autocomplete="one-time-code"
//
// If not found, then we also check for:
//
// - Inputs that have type="password" attribute
// - Inputs that have type="email" attribute
// - Form with defined action and method
//
// Cases that aren't covered:
// - multipage (requires JS or some insane parsing)
// - protected login forms (security through obscurity).
func HasLoginForm(doc *goquery.Document) bool {
	// At minimum, we expect either:
	// - autocomplete hints
	// - OR a password field with an email field AND a form with method+action
	return hasAutoCompleteHints(doc) || (hasValidForm(doc) && hasPasswordInput(doc) && hasLoginInput(doc))
}

// Reference: https://developer.1password.com/docs/web/compatible-website-design/
func hasAutoCompleteHints(doc *goquery.Document) bool {
	for _, attr := range CommonAutocompleteValues {
		if doc.Find(fmt.Sprintf("[autocomplete='%s']", attr)).Length() > 0 {
			return true
		}
	}

	return false
}

func hasPasswordInput(doc *goquery.Document) bool {
	return doc.Find(`input[type='password']`).Length() > 0
}

func hasLoginInput(doc *goquery.Document) bool {
	for _, attr := range CommonLoginInputValues {
		if doc.Find(fmt.Sprintf("[type='%s']", attr)).Length() > 0 {
			return true
		}
	}

	return false
}

func hasValidForm(doc *goquery.Document) bool {
	found := false

	doc.Find("form").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		method, hasMethod := s.Attr("method")

		action, hasAction := s.Attr("action")

		if hasMethod && hasAction && strings.TrimSpace(method) != "" && strings.TrimSpace(action) != "" {
			found = true

			return false
		}

		return true
	})

	return found
}
