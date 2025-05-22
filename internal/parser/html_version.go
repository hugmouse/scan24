package parser

import (
	"errors"
	"strings"
	"unicode"
)

// DoctypeNode holds the parsed DOCTYPE information
//
// Read more about identifiers:
// https://en.wikipedia.org/wiki/Document_type_declaration#External_identifier
type DoctypeNode struct {
	Name             string  // e.g. "HTML5", "HTML 4.01 Strict", "XHTML 1.0 Transitional"
	DocumentTypeName string  // the root name, always lower-case ("html")
	PublicID         *string // the PUBLIC identifier, OPTIONAL
	SystemID         *string // the SYSTEM identifier/URI, OPTIONAL
	Raw              string  // the exact "<!DOCTYPE ...>" string parsed
}

// KnownDOCTYPEs are the supported DOCTYPE formats,
// excluding XHTML5 and HTML5 (handled in GetHTMLVersion explicitly)
var KnownDOCTYPEs = map[string]string{
	"-//W3C//DTD HTML 2.0//EN":                 "HTML 2.0",
	"-//IETF//DTD HTML 2.0//EN":                "HTML 2.0",
	"-//W3C//DTD HTML 3.2 Final//EN":           "HTML 3.2",
	"-//IETF//DTD HTML 3.2 Final//EN":          "HTML 3.2",
	"-//W3C//DTD HTML 4.0//EN":                 "HTML 4.0 Strict",
	"-//W3C//DTD HTML 4.0 Transitional//EN":    "HTML 4.0 Transitional",
	"-//W3C//DTD HTML 4.0 Frameset//EN":        "HTML 4.0 Frameset",
	"-//W3C//DTD HTML 4.01//EN":                "HTML 4.01 Strict",
	"-//IETF//DTD HTML 4.01//EN":               "HTML 4.01 Strict",
	"-//W3C//DTD HTML 4.01 Transitional//EN":   "HTML 4.01 Transitional",
	"-//IETF//DTD HTML 4.01 Transitional//EN":  "HTML 4.01 Transitional",
	"-//W3C//DTD HTML 4.01 Frameset//EN":       "HTML 4.01 Frameset",
	"-//IETF//DTD HTML 4.01 Frameset//EN":      "HTML 4.01 Frameset",
	"-//W3C//DTD XHTML 1.0 Strict//EN":         "XHTML 1.0 Strict",
	"-//IETF//DTD XHTML 1.0 Strict//EN":        "XHTML 1.0 Strict",
	"-//W3C//DTD XHTML 1.0 Transitional//EN":   "XHTML 1.0 Transitional",
	"-//IETF//DTD XHTML 1.0 Transitional//EN":  "XHTML 1.0 Transitional",
	"-//W3C//DTD XHTML 1.0 Frameset//EN":       "XHTML 1.0 Frameset",
	"-//IETF//DTD XHTML 1.0 Frameset//EN":      "XHTML 1.0 Frameset",
	"-//W3C//DTD XHTML 1.1//EN":                "XHTML 1.1",
	"-//IETF//DTD XHTML 1.1//EN":               "XHTML 1.1",
	"-//W3C//DTD XHTML Basic 1.1//EN":          "XHTML Basic 1.1",
	"-//W3C//DTD XHTML-Print 1.0//EN":          "XHTML-Print 1.0",
	"-//W3C//DTD XHTML Basic 1.0//EN":          "XHTML Basic 1.0",
	"-//IETF//DTD XHTML Basic 1.0//EN":         "XHTML Basic 1.0",
	"-//IETF//DTD HTML 2.0 Level 1//EN":        "HTML 2.0 Level 1",
	"-//IETF//DTD HTML 2.0 Level 2//EN":        "HTML 2.0 Level 2",
	"-//IETF//DTD HTML 2.0 Strict//EN":         "HTML 2.0 Strict",
	"-//IETF//DTD HTML 2.0 Strict Level 1//EN": "HTML 2.0 Strict Level 1",
	"-//IETF//DTD HTML 2.0 Strict Level 2//EN": "HTML 2.0 Strict Level 2",
	"-//W3C//DTD HTML 3.0//EN":                 "HTML 3.0 Draft",
	"-//IETF//DTD HTML 3.0//EN":                "HTML 3.0 Draft",
	"-//IETF//DTD HTML Level 3//EN":            "HTML Level 3",
	"-//IETF//DTD HTML Strict Level 3//EN":     "HTML Strict Level 3",
	"-//W3C//DTD MathML 2.0//EN":               "MathML 2.0",
}

// GetHTMLVersion parses out the DOCTYPE and classifies it.  No regex.
func GetHTMLVersion(input string) (*DoctypeNode, error) {
	raw := input
	up := strings.ToUpper(input)

	// 1) Find the DOCTYPE
	startIndex := strings.Index(up, "<!DOCTYPE")
	if startIndex < 0 {
		// <!DOCTYPE is not present, either that a quirky document or XHTML5
		//
		// We need to do additional check for having <HTML next to determine that
		if strings.Contains(strings.ToUpper(input), "<HTML") {
			return &DoctypeNode{
				Name:             "XHTML5",
				DocumentTypeName: "html",
				Raw:              "",
			}, nil
		}
		return nil, errors.New("unsupported DOCTYPE or quirky document (no <!DOCTYPE>)")
	}

	// Now we need to find `>` for the `<!DOCTYPE`
	endIndex := strings.IndexRune(up[startIndex:], '>')
	if endIndex < 0 {
		return nil, errors.New("malformed <!DOCTYPE> (no closing '>')")
	}
	rawDoctype := raw[startIndex : startIndex+endIndex+1]

	// 2) Tokenize between `<!DOCTYPE` and `>`
	//
	// Splits on whitespace except inside quoted strings
	// TODO: check for ASCII hyphen case (https://en.wikipedia.org/wiki/Document_type_declaration#Document_type_name)
	//
	// Preserves the exact PUBLIC/SYSTEM identifier strings
	var tokens []string
	i := startIndex + len("<!DOCTYPE")
	for i < startIndex+endIndex {

		// Skip whitespace
		for i < startIndex+endIndex && unicode.IsSpace(rune(raw[i])) {
			i++
		}

		// We are at the end, kill it!
		if i >= startIndex+endIndex {
			break
		}

		// Quoted string case
		if raw[i] == '"' {
			j := i + 1
			for j < startIndex+endIndex && raw[j] != '"' {
				j++
			}
			if j >= startIndex+endIndex {
				return nil, errors.New("unterminated quote in <!DOCTYPE>")
			}
			tokens = append(tokens, raw[i+1:j])
			i = j + 1
		} else {
			// Unquoted word
			j := i
			for j < startIndex+endIndex && !unicode.IsSpace(rune(raw[j])) {
				j++
			}
			tokens = append(tokens, strings.ToUpper(raw[i:j]))
			i = j
		}
	}

	// tokens[0] is ALWAYS the root name, unless malformed: "html", "math", "svg"
	if len(tokens) == 0 {
		return nil, errors.New("DOCTYPE has no name token, document considered quirky")
	}
	root := strings.ToLower(tokens[0])

	// 3) Match based on tokens
	// a) <!DOCTYPE html> -> HTML5
	if len(tokens) == 1 && root == "html" {
		return &DoctypeNode{
			Name:             "HTML5",
			DocumentTypeName: root,
			Raw:              rawDoctype,
		}, nil
	}

	// b) PUBLIC + System form: ["HTML","PUBLIC", publicId, systemId]
	if len(tokens) >= 4 && tokens[1] == "PUBLIC" {
		pub := tokens[2]
		sys := tokens[3]
		name, known := KnownDOCTYPEs[pub]
		if !known {
			name = "Unknown PUBLIC+SYSTEM declaration"
		}
		return &DoctypeNode{
			Name:             name,
			DocumentTypeName: root,
			PublicID:         &pub,
			SystemID:         &sys,
			Raw:              rawDoctype,
		}, nil
	}

	// c) PUBLIC-only form: ["HTML", "PUBLIC", publicId]
	if len(tokens) >= 3 && tokens[1] == "PUBLIC" {
		pub := tokens[2]
		name, known := KnownDOCTYPEs[pub]
		if !known {
			name = "Unknown PUBLIC-only declaration"
		}
		return &DoctypeNode{
			Name:             name,
			DocumentTypeName: root,
			PublicID:         &pub,
			Raw:              rawDoctype,
		}, nil
	}

	// d) SYSTEM-only form: ["HTML","SYSTEM", systemId]
	if len(tokens) >= 3 && tokens[1] == "SYSTEM" {
		sys := tokens[2]
		return &DoctypeNode{
			Name:             "Unknown SYSTEM-only declaration",
			DocumentTypeName: root,
			SystemID:         &sys,
			Raw:              rawDoctype,
		}, nil
	}

	// 4) Bare <!DOCTYPE i_hate_standards> (quirky) or unrecognized
	return nil, errors.New("unsupported or unrecognized DOCTYPE format, document considered quirky")
}
