package test

import (
	"github.com/hugmouse/scan24/internal/parser"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGetHTMLVersion(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"html2.0.html", "HTML 2.0"},
		{"html3.2.html", "HTML 3.2"},
		{"html4.01-strict.html", "HTML 4.01 Strict"},
		{"html4.01-transitional.html", "HTML 4.01 Transitional"},
		{"html4.01-frameset.html", "HTML 4.01 Frameset"},
		{"html5.html", "HTML5"},
		{"xhtml1.0-strict.html", "XHTML 1.0 Strict"},
		{"xhtml1.0-transitional.html", "XHTML 1.0 Transitional"},
		{"xhtml1.0-frameset.html", "XHTML 1.0 Frameset"},
		{"xhtml1.1.html", "XHTML 1.1"},
		{"xhtml5.html", "XHTML5"},
		{"xhtml-basic-1.0.html", "XHTML Basic 1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			path := filepath.Join("testdata", "doctype", tt.filename)

			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("Failed to open file %s: %v", tt.filename, err)
			}

			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", tt.filename, err)
			}

			actual, err := parser.GetHTMLVersion(string(bytes))
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.filename, err)
			}

			if actual.Name != tt.expected {
				t.Errorf("GetHTMLVersion(%s) = %q; want %q", tt.filename, actual.Name, tt.expected)
			}
		})
	}
}
