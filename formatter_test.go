package main

import (
	"bytes"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"testing/quick"
	"text/template"
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
			name:   "spacing preserves negative number without trim marker whitespace",
			input:  "{{-3}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ -3 }}\n",
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
			name:   "spacing ignores close delimiter in quoted string",
			input:  `{{printf "%s" "}}"}}` + "\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: `{{ printf "%s" "}}" }}` + "\n",
		},
		{
			name:   "spacing ignores close delimiter in raw string",
			input:  "{{printf `}}`}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ printf `}}` }}\n",
		},
		{
			name:   "spacing preserves comment containing close delimiter",
			input:  "{{/* }} */}}\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{/* }} */}}\n",
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
			name:   "preserves missing trailing newline",
			input:  "{{.foo}}",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: "{{ .foo }}",
		},
		{
			name:   "long line",
			input:  strings.Repeat("a", 70_000) + "\n",
			opts:   Options{Spacing: true, IndentStyle: "none"},
			expect: strings.Repeat("a", 70_000) + "\n",
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

func TestFormatSpacingPreservesIntegerActionOutput(t *testing.T) {
	err := quick.Check(func(n int) bool {
		input := "{{" + strconv.Itoa(n) + "}}"

		formatted, err := formatString(input, Options{Spacing: true, IndentStyle: "none"})
		if err != nil {
			t.Logf("format %q: %v", input, err)
			return false
		}

		before, err := executeTemplate(input)
		if err != nil {
			t.Logf("execute input %q: %v", input, err)
			return false
		}
		after, err := executeTemplate(formatted)
		if err != nil {
			t.Logf("execute formatted %q from %q: %v", formatted, input, err)
			return false
		}
		if before != after {
			t.Logf("format changed output: input %q -> %q, before %q, after %q", input, formatted, before, after)
			return false
		}
		return true
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatNoOptionsIsIdentity(t *testing.T) {
	err := quick.Check(func(input string) bool {
		formatted, err := formatString(input, Options{IndentStyle: "none"})
		if err != nil {
			t.Logf("format %q: %v", input, err)
			return false
		}
		if formatted != input {
			t.Logf("format changed input with no options: input %q -> %q", input, formatted)
			return false
		}
		return true
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatIsIdempotent(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{name: "none", opts: Options{IndentStyle: "none"}},
		{name: "spacing", opts: Options{Spacing: true, IndentStyle: "none"}},
		{name: "spaces", opts: Options{IndentStyle: "spaces", IndentSize: 2}},
		{name: "tabs", opts: Options{IndentStyle: "tabs", IndentSize: 1}},
		{name: "spacing and spaces", opts: Options{Spacing: true, IndentStyle: "spaces", IndentSize: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := quick.Check(func(input string) bool {
				once, err := formatString(input, tt.opts)
				if err != nil {
					t.Logf("format once %q with %+v: %v", input, tt.opts, err)
					return false
				}
				twice, err := formatString(once, tt.opts)
				if err != nil {
					t.Logf("format twice %q with %+v: %v", once, tt.opts, err)
					return false
				}
				if twice != once {
					t.Logf("format not idempotent with %+v: input %q -> %q -> %q", tt.opts, input, once, twice)
					return false
				}
				return true
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFormatSpacingPreservesStringLiteralActionOutput(t *testing.T) {
	err := quick.Check(func(s string) bool {
		input := "{{" + strconv.Quote(s) + "}}"

		formatted, err := formatString(input, Options{Spacing: true, IndentStyle: "none"})
		if err != nil {
			t.Logf("format %q: %v", input, err)
			return false
		}

		before, err := executeTemplate(input)
		if err != nil {
			t.Logf("execute input %q: %v", input, err)
			return false
		}
		after, err := executeTemplate(formatted)
		if err != nil {
			t.Logf("execute formatted %q from %q: %v", formatted, input, err)
			return false
		}
		if before != after {
			t.Logf("format changed output: input %q -> %q, before %q, after %q", input, formatted, before, after)
			return false
		}
		return true
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolveVersion(t *testing.T) {
	originalVersion := version
	t.Cleanup(func() {
		version = originalVersion
	})

	tests := []struct {
		name        string
		version     string
		buildInfo   *debug.BuildInfo
		buildInfoOK bool
		want        string
	}{
		{
			name:        "ldflags version wins",
			version:     "v1.2.3",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}},
			buildInfoOK: true,
			want:        "v1.2.3",
		},
		{
			name:        "build info version",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			buildInfoOK: true,
			want:        "v1.2.3",
		},
		{
			name:        "devel build info falls back to devel",
			buildInfo:   &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			buildInfoOK: true,
			want:        "(devel)",
		},
		{
			name:        "empty build info version falls back to devel",
			buildInfo:   &debug.BuildInfo{},
			buildInfoOK: true,
			want:        "(devel)",
		},
		{
			name: "missing build info falls back to devel",
			want: "(devel)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version = tt.version
			got := resolveVersion(func() (*debug.BuildInfo, bool) {
				return tt.buildInfo, tt.buildInfoOK
			})
			if got != tt.want {
				t.Fatalf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func formatString(input string, opts Options) (string, error) {
	var buf bytes.Buffer
	if err := Format(&buf, strings.NewReader(input), opts); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func executeTemplate(src string) (string, error) {
	tmpl, err := template.New("test").Parse(src)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}
