package media

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

// DescriptionsFile is where the descriptions people write for screen readers
// are kept. They belong to the site rather than to ChatCMS, so they live in
// the site and travel with it; a picture's own bytes have nowhere to put them.
const DescriptionsFile = "quiqr/media.yaml"

type descriptions struct {
	Descriptions map[string]string `yaml:"descriptions"`
}

// ReadDescriptions reads what has been written for each picture. A site
// without the file has written none, which is not a failure.
func ReadDescriptions(fsys fs.FS) map[string]string {
	content, err := fs.ReadFile(fsys, DescriptionsFile)
	if err != nil {
		return map[string]string{}
	}

	var held descriptions
	if err := yaml.Unmarshal(content, &held); err != nil || held.Descriptions == nil {
		return map[string]string{}
	}
	return held.Descriptions
}

// WriteDescriptions gives back the file to save, with the names in order so
// that the same descriptions always make the same file.
func WriteDescriptions(held map[string]string) ([]byte, error) {
	kept := map[string]string{}
	names := make([]string, 0, len(held))
	for name, said := range held {
		if strings.TrimSpace(said) == "" {
			continue
		}
		kept[name] = said
		names = append(names, name)
	}
	sort.Strings(names)

	out := strings.Builder{}
	out.WriteString("# What each picture shows, for people using a screen reader.\n")
	out.WriteString("descriptions:\n")
	for _, name := range names {
		encoded, err := yaml.Marshal(kept[name])
		if err != nil {
			return nil, err
		}
		out.WriteString("  " + name + ": " + strings.TrimRight(string(encoded), "\n") + "\n")
	}
	return []byte(out.String()), nil
}
