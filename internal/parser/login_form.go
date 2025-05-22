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
// - protected login forms (security through obscurity)
func HasLoginForm(doc *goquery.Document) bool {
	// 1. Check for elements with autocomplete attributes used by password managers
	//
	// Reference: https://developer.1password.com/docs/web/compatible-website-design/
	foundAutoComplete := false
	for _, attr := range CommonAutocompleteValues {
		if doc.Find(fmt.Sprintf("[autocomplete='%s']", attr)).Length() > 0 {
			foundAutoComplete = true
			break
		}
	}

	// 2. Check for password inputs
	passwordInputs := doc.Find(`input[type='password']`)
	foundPasswordInput := passwordInputs.Length() > 0

	// 3. Check for login inputs
	var foundEmailInput bool
	for _, attr := range CommonLoginInputValues {
		if doc.Find(fmt.Sprintf("[type='%s']", attr)).Length() > 0 {
			foundEmailInput = true
			break
		}
	}

	// 4. Check for forms with both method and action defined
	hasFormWithMethodAndAction := false
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		method, existsMethod := s.Attr("method")
		action, existsAction := s.Attr("action")

		if existsMethod && existsAction && strings.TrimSpace(method) != "" && strings.TrimSpace(action) != "" {
			hasFormWithMethodAndAction = true
		}
	})

	// At minimum, we expect either:
	// - autocomplete hints
	// - OR a password field with (optionally) an email field AND a form with method+action
	return foundAutoComplete || (hasFormWithMethodAndAction && foundPasswordInput && foundEmailInput)
}
