package main

import (
	"bufio"
	"io"
	"strings"
)

type Options struct {
	Spacing     bool
	IndentStyle string // "none", "tabs", "spaces"
	IndentSize  int
}

func Format(w io.Writer, r io.Reader, opts Options) error {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if opts.Spacing {
		for i, line := range lines {
			lines[i] = applySpacing(line)
		}
	}

	if opts.IndentStyle != "none" && opts.IndentStyle != "" {
		lines = applyIndent(lines, opts)
	}

	for i, line := range lines {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	// Preserve trailing newline if input had content
	if len(lines) > 0 {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

func applySpacing(line string) string {
	var b strings.Builder
	rest := line
	for {
		openIdx := strings.Index(rest, "{{")
		if openIdx == -1 {
			b.WriteString(rest)
			break
		}
		closeIdx := strings.Index(rest[openIdx:], "}}")
		if closeIdx == -1 {
			b.WriteString(rest)
			break
		}
		closeIdx += openIdx

		// Write text before the action
		b.WriteString(rest[:openIdx])

		action := rest[openIdx : closeIdx+2]
		b.WriteString(reformatAction(action))

		rest = rest[closeIdx+2:]
	}
	return b.String()
}

func reformatAction(action string) string {
	inner := action[2 : len(action)-2]

	// Detect comment
	trimmed := strings.TrimSpace(inner)
	if strings.HasPrefix(trimmed, "/*") {
		return action
	}

	// Detect trim markers
	openDelim := "{{"
	closeDelim := "}}"

	body := inner

	if strings.HasPrefix(strings.TrimLeft(body, " \t"), "-") {
		openDelim = "{{-"
		body = strings.TrimLeft(body, " \t")
		body = body[1:] // remove the '-'
	}

	if strings.HasSuffix(strings.TrimRight(body, " \t"), "-") {
		closeDelim = "-}}"
		body = strings.TrimRight(body, " \t")
		body = body[:len(body)-1] // remove the '-'
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return openDelim + " " + closeDelim
	}

	return openDelim + " " + body + " " + closeDelim
}

func applyIndent(lines []string, opts Options) []string {
	var unit string
	switch opts.IndentStyle {
	case "tabs":
		unit = strings.Repeat("\t", opts.IndentSize)
	case "spaces":
		unit = strings.Repeat(" ", opts.IndentSize)
	default:
		return lines
	}

	level := 0
	result := make([]string, len(lines))

	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			result[i] = ""
			continue
		}

		keyword := extractKeyword(trimmed)

		switch keyword {
		case "end":
			level = max(level-1, 0)
			result[i] = strings.Repeat(unit, level) + trimmed
		case "else":
			result[i] = strings.Repeat(unit, max(level-1, 0)) + trimmed
		case "range", "if", "with", "define", "block":
			result[i] = strings.Repeat(unit, level) + trimmed
			level++
		default:
			result[i] = strings.Repeat(unit, level) + trimmed
		}
	}

	return result
}

func extractKeyword(line string) string {
	openIdx := strings.Index(line, "{{")
	if openIdx != 0 {
		return ""
	}
	closeIdx := strings.Index(line, "}}")
	if closeIdx == -1 {
		return ""
	}

	inner := line[2:closeIdx]

	// Strip trim marker
	inner = strings.TrimSpace(inner)
	inner = strings.TrimPrefix(inner, "-")
	inner = strings.TrimSpace(inner)

	fields := strings.Fields(inner)
	if len(fields) == 0 {
		return ""
	}

	keyword := fields[0]
	// Handle "else if"
	if keyword == "else" && len(fields) > 1 && fields[1] == "if" {
		return "else"
	}
	return keyword
}
