package test

import (
	"github.com/hugmouse/scan24/internal/parser"
	"testing"
)

func TestClassify(t *testing.T) {
	type args struct {
		href string
	}

	tests := []struct {
		name string
		args args
		want parser.HrefType
	}{
		{name: "Internal (#)", args: args{href: "#"}, want: parser.Internal},
		{name: "Internal (!#)", args: args{href: "!#"}, want: parser.Internal},
		{name: "Internal (.)", args: args{href: "."}, want: parser.Internal},
		{name: "Internal (..)", args: args{href: ".."}, want: parser.Internal},
		{name: "Internal (./)", args: args{href: "./"}, want: parser.Internal},
		{name: "Internal (:)", args: args{href: ":"}, want: parser.Internal},
		{name: "Internal (example.com)", args: args{href: "example.com"}, want: parser.Internal},
		{name: "External (http://example.com)", args: args{href: "http://example.com"}, want: parser.External},
		{name: "External (https://example.com)", args: args{href: "https://example.com"}, want: parser.External},
		{name: "External (ftp://example.com)", args: args{href: "ftp://example.com"}, want: parser.External},
		{name: "Protocol (tel:)", args: args{href: "tel:"}, want: parser.Protocol},
		{name: "Protocol non-standard (steam:)", args: args{href: "steam:"}, want: parser.Protocol},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Classify(tt.args.href); got != tt.want {
				t.Errorf("Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}
