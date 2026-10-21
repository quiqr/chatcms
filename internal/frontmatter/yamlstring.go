package frontmatter

import (
	"fmt"
	"regexp"
	"strings"
)

// yamlString chooses between the styles js-yaml chooses between: plain,
// single quoted, double quoted, and a literal or folded block. The rules here
// were derived by running js-yaml over the table in scripts/goldens, not from
// reading its source, and every one of them is pinned by a golden file.
func yamlString(value string, blockIndent int, allowBlock bool) (string, error) {
	if value == "" {
		return "''", nil
	}

	if needsDoubleQuotes(value) {
		return doubleQuote(value), nil
	}

	if strings.Contains(value, "\n") {
		if !allowBlock {
			return singleQuote(value), nil
		}
		return literalBlock(value, blockIndent), nil
	}

	if plainAllowed(value) {
		width := yamlLineWidth - blockIndent
		if len(value) > width && allowBlock {
			return foldedBlock(value, blockIndent, width), nil
		}
		return value, nil
	}

	return singleQuote(value), nil
}

// needsDoubleQuotes covers the characters a single-quoted scalar cannot
// carry, which js-yaml escapes instead.
func needsDoubleQuotes(value string) bool {
	for _, r := range value {
		switch r {
		case '\n':
			continue
		case '\t', '\r', '\v', '\f', '\b', 0x85, 0x2028, 0x2029:
			return true
		}
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

var doubleQuoteEscapes = map[rune]string{
	'"':  `\"`,
	'\\': `\\`,
	'\n': `\n`,
	'\t': `\t`,
	'\r': `\r`,
	'\v': `\v`,
	'\f': `\f`,
	'\b': `\b`,
	0x00: `\0`,
	0x1b: `\e`,
	0x85: `\N`,
	0xa0: `\_`,
}

func doubleQuote(value string) string {
	var out strings.Builder
	out.WriteByte('"')

	for _, r := range value {
		if escape, ok := doubleQuoteEscapes[r]; ok {
			out.WriteString(escape)
			continue
		}
		if r < 0x20 || r == 0x7f {
			fmt.Fprintf(&out, `\x%02x`, r)
			continue
		}
		out.WriteRune(r)
	}

	out.WriteByte('"')
	return out.String()
}

func singleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// literalBlock writes a multi-line string as a block, choosing the chomping
// indicator from how the string ends: none for exactly one trailing newline,
// a minus for none, a plus for more than one.
func literalBlock(value string, indent int) string {
	body := value
	indicator := ""

	switch {
	case !strings.HasSuffix(body, "\n"):
		indicator = "-"
		body += "\n"
	case strings.HasSuffix(body, "\n\n"):
		indicator = "+"
	}

	prefix := strings.Repeat(" ", indent)
	var out strings.Builder
	out.WriteString("|")
	out.WriteString(indicator)
	out.WriteString("\n")

	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	for _, line := range lines {
		if line == "" {
			out.WriteString("\n")
			continue
		}
		out.WriteString(prefix)
		out.WriteString(line)
		out.WriteString("\n")
	}
	return strings.TrimSuffix(out.String(), "\n")
}

// foldedBlock wraps a single-line string across lines at the first space at
// or before the width, which is what js-yaml does. A run with no space in it
// is left long rather than broken mid-word.
func foldedBlock(value string, indent, width int) string {
	prefix := strings.Repeat(" ", indent)

	var out strings.Builder
	out.WriteString(">-\n")

	for _, line := range foldLines(value, width) {
		out.WriteString(prefix)
		out.WriteString(line)
		out.WriteString("\n")
	}
	return strings.TrimSuffix(out.String(), "\n")
}

func foldLines(value string, width int) []string {
	var lines []string
	rest := value

	for len(rest) > width {
		cut := strings.LastIndex(rest[:width+1], " ")
		if cut <= 0 {
			// No break opportunity within the width: take the next one, or
			// the whole remainder when there is none.
			next := strings.Index(rest, " ")
			if next < 0 {
				break
			}
			cut = next
		}
		lines = append(lines, rest[:cut])
		rest = rest[cut+1:]
	}

	return append(lines, rest)
}

var (
	// A plain scalar may not start with an indicator, may not contain a colon
	// followed by a space or a space followed by a hash, and may not have
	// leading or trailing whitespace.
	plainLeadingIndicator = regexp.MustCompile(`^[-?:,\[\]{}#&*!|>'"%@` + "`" + `]`)

	// Anything js-yaml's resolvers would read back as another type has to be
	// quoted, or the value would change meaning on the way back in.
	resolvesToNull  = regexp.MustCompile(`^(?:~|null|Null|NULL)$`)
	resolvesToBool  = regexp.MustCompile(`^(?:true|True|TRUE|false|False|FALSE|yes|Yes|YES|no|No|NO|on|On|ON|off|Off|OFF|y|Y|n|N)$`)
	resolvesToInt   = regexp.MustCompile(`^[-+]?(?:0b[01_]+|0o?[0-7_]+|0x[0-9a-fA-F_]+|(?:[0-9][0-9_]*)(?::[0-5]?[0-9])*)$`)
	resolvesToFloat = regexp.MustCompile(`^[-+]?(?:` +
		`[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*` +
		`|[-+]?\.(?:inf|Inf|INF)` +
		`|[0-9][0-9_]*\.[0-9_]*(?:[eE][-+]?[0-9]+)?` +
		`|\.[0-9_]+(?:[eE][-+]?[0-9]+)?` +
		`)$`)
	resolvesToNaN = regexp.MustCompile(`^\.(?:nan|NaN|NAN)$`)
)

func plainAllowed(value string) bool {
	if value == "" {
		return false
	}
	if strings.TrimSpace(value) != value {
		return false
	}
	if plainLeadingIndicator.MatchString(value) {
		return false
	}
	if strings.Contains(value, ": ") || strings.HasSuffix(value, ":") {
		return false
	}
	if strings.Contains(value, " #") {
		return false
	}
	if strings.Contains(value, "\n") {
		return false
	}
	if value == "---" || value == "..." {
		return false
	}
	return !resolvesToOtherType(value)
}

func resolvesToOtherType(value string) bool {
	switch {
	case resolvesToNull.MatchString(value),
		resolvesToBool.MatchString(value),
		resolvesToInt.MatchString(value),
		resolvesToFloat.MatchString(value),
		resolvesToNaN.MatchString(value),
		looksLikeTimestamp(value):
		return true
	}
	return false
}

func join(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
