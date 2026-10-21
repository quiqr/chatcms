package frontmatter

import (
	"fmt"
	"strings"
)

// Bytes writes a document back as a content file. Only YAML front matter is
// written; see ErrUnwritableFormat and the assumption in the change that
// introduced this package.
func (d *Document) Bytes() ([]byte, error) {
	if d.Fields == nil || d.Fields.Len() == 0 {
		return []byte(d.Body), nil
	}

	switch d.Format {
	case FormatYAML, FormatNone:
	case FormatTOML, FormatJSON:
		return nil, fmt.Errorf("%w: %s", ErrUnwritableFormat, d.Format)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnwritableFormat, d.Format)
	}

	header, err := EmitYAML(d.Fields)
	if err != nil {
		return nil, err
	}

	var out strings.Builder
	out.WriteString(yamlDelimiter)
	out.WriteString("\n")
	out.Write(header)
	out.WriteString(yamlDelimiter)
	out.WriteString("\n\n")
	out.WriteString(d.Body)
	return []byte(out.String()), nil
}
