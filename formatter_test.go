package main

import (
	"bytes"
	"strings"
	"testing"
)

func lines(ss ...string) string {
	return strings.Join(ss, "\n") + "\n"
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		opts   Options
		expect string
	}{
		// Spacing tests
		{
			name:   "spacing basic",
			input:  "{{.foo}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .foo }}\n",
		},
		{
			name:   "spacing already spaced",
			input:  "{{ .foo }}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .foo }}\n",
		},
		{
			name:   "spacing with keyword",
			input:  "{{range .foo}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ range .foo }}\n",
		},
		{
			name:   "spacing trim markers both",
			input:  "{{- .foo -}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{- .foo -}}\n",
		},
		{
			name:   "spacing trim markers no space",
			input:  "{{-.foo-}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{- .foo -}}\n",
		},
		{
			name:   "spacing trim marker left only",
			input:  "{{- .foo}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{- .foo }}\n",
		},
		{
			name:   "spacing trim marker right only",
			input:  "{{.foo -}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .foo -}}\n",
		},
		{
			name:   "spacing comment passthrough",
			input:  "{{/* comment */}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{/* comment */}}\n",
		},
		{
			name:   "spacing multiple actions",
			input:  "{{.a}}{{.b}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .a }}{{ .b }}\n",
		},
		{
			name:   "spacing pipeline",
			input:  "{{.foo | bar}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .foo | bar }}\n",
		},
		{
			name:   "spacing disabled",
			input:  "{{.foo}}\n",
			opts:   Options{Spacing: false, IndentStyle: "none"},
			expect: "{{.foo}}\n",
		},
		// Indent tests
		{
			name: "indent range/end spaces",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{range .foo}}",
				"  {{.bar}}",
				"{{end}}",
			),
		},
		{
			name: "indent if/else/end",
			input: lines(
				"{{if .x}}",
				"{{.a}}",
				"{{else}}",
				"{{.b}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{if .x}}",
				"  {{.a}}",
				"{{else}}",
				"  {{.b}}",
				"{{end}}",
			),
		},
		{
			name: "indent if/else if/else/end",
			input: lines(
				"{{if .x}}",
				"{{.a}}",
				"{{else if .y}}",
				"{{.b}}",
				"{{else}}",
				"{{.c}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{if .x}}",
				"  {{.a}}",
				"{{else if .y}}",
				"  {{.b}}",
				"{{else}}",
				"  {{.c}}",
				"{{end}}",
			),
		},
		{
			name: "indent nested",
			input: lines(
				"{{range .items}}",
				"{{if .visible}}",
				"{{.name}}",
				"{{end}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{range .items}}",
				"  {{if .visible}}",
				"    {{.name}}",
				"  {{end}}",
				"{{end}}",
			),
		},
		{
			name: "indent with",
			input: lines(
				"{{with .ctx}}",
				"{{.val}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{with .ctx}}",
				"  {{.val}}",
				"{{end}}",
			),
		},
		{
			name: "indent define",
			input: lines(
				`{{define "tmpl"}}`,
				"{{.content}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				`{{define "tmpl"}}`,
				"  {{.content}}",
				"{{end}}",
			),
		},
		{
			name: "indent block",
			input: lines(
				`{{block "tmpl" .}}`,
				"{{.content}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				`{{block "tmpl" .}}`,
				"  {{.content}}",
				"{{end}}",
			),
		},
		{
			name: "indent tabs",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "tabs", IndentSize: 1},
			expect: lines(
				"{{range .foo}}",
				"\t{{.bar}}",
				"{{end}}",
			),
		},
		{
			name: "indent size 4",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 4},
			expect: lines(
				"{{range .foo}}",
				"    {{.bar}}",
				"{{end}}",
			),
		},
		{
			name: "indent preserves empty lines",
			input: lines(
				"{{range .foo}}",
				"",
				"{{.bar}}",
				"{{end}}",
			),
			opts:   Options{IndentStyle: "spaces", IndentSize: 2},
			expect: "{{range .foo}}\n\n  {{.bar}}\n{{end}}\n",
		},
		{
			name: "indent none",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "none", IndentSize: 2},
			expect: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
		},
		// Combined tests
		{
			name: "spacing and indent combined",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{Spacing: true, IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{ range .foo }}",
				"  {{ .bar }}",
				"{{ end }}",
			),
		},
		{
			name: "readme example",
			input: lines(
				"{{range .foo}}",
				"{{.bar}}",
				"{{end}}",
			),
			opts: Options{Spacing: true, IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{ range .foo }}",
				"  {{ .bar }}",
				"{{ end }}",
			),
		},
		// Edge cases
		{
			name:   "empty input",
			input:  "",
			opts:   Options{Spacing: true, IndentStyle: "spaces", IndentSize: 2},
			expect: "",
		},
		{
			name:   "no actions",
			input:  "hello world\n",
			opts:   Options{Spacing: true, IndentStyle: "spaces", IndentSize: 2},
			expect: "hello world\n",
		},
		{
			name:   "text with actions inline",
			input:  "<p>{{.name}}</p>\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "<p>{{ .name }}</p>\n",
		},
		{
			name: "indent strips existing indentation",
			input: lines(
				"{{range .foo}}",
				"    {{.bar}}",
				"{{end}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{range .foo}}",
				"  {{.bar}}",
				"{{end}}",
			),
		},
		{
			name: "indent with trim markers",
			input: lines(
				"{{- range .foo -}}",
				"{{- .bar -}}",
				"{{- end -}}",
			),
			opts: Options{IndentStyle: "spaces", IndentSize: 2},
			expect: lines(
				"{{- range .foo -}}",
				"  {{- .bar -}}",
				"{{- end -}}",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Format(&buf, strings.NewReader(tt.input), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if got != tt.expect {
				t.Errorf("mismatch\ninput:  %q\ngot:    %q\nexpect: %q", tt.input, got, tt.expect)
			}
		})
	}
}
