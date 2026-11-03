// Command reference answers questions about a site mounted in the reference
// instance: what its model declares, and where a single keeps its file. It
// reads the model with the same parser the server uses, so a model the
// reference can be driven against but ChatCMS cannot read says so here,
// before anybody records an answer from it.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/quiqr/chatcms/internal/model"
)

func main() {
	if len(os.Args) < 3 {
		fail("usage: reference <fields|single> <site> [single]")
	}

	site := os.DirFS(os.Args[2])
	loaded, err := model.Load(site)
	if err != nil {
		fail("the model could not be read: %v", err)
	}

	switch os.Args[1] {
	case "fields":
		fields(loaded)
	case "single":
		if len(os.Args) < 4 {
			fail("usage: reference single <site> <single>")
		}
		single(loaded, os.Args[3])
	default:
		fail("usage: reference <fields|single> <site> [single]")
	}
}

func fields(loaded *model.Model) {
	counted := map[string]int{}
	for _, one := range loaded.Singles {
		count(one.Fields, counted)
	}
	for _, one := range loaded.Collections {
		count(one.Fields, counted)
	}

	types := make([]string, 0, len(counted))
	for kind := range counted {
		types = append(types, kind)
	}
	sort.Slice(types, func(i, j int) bool {
		if counted[types[i]] != counted[types[j]] {
			return counted[types[i]] > counted[types[j]]
		}
		return types[i] < types[j]
	})

	report := map[string]any{
		"singles":     len(loaded.Singles),
		"collections": len(loaded.Collections),
		"types":       counted,
		"order":       types,
		"warnings":    loaded.Warnings,
	}
	write(report)
}

func count(fields []model.Field, into map[string]int) {
	for _, field := range fields {
		into[field.Type]++
		count(field.Fields, into)
	}
}

// single says where a single keeps its file and what the old application
// calls it on screen. The title matters because the old application's own
// routes answer 404 to a browser asking for one directly: a form is reached
// by clicking its name, so a recording has to know that name.
func single(loaded *model.Model, key string) {
	for _, one := range loaded.Singles {
		if one.Key != key {
			continue
		}
		if one.File == "" {
			fail("the single %q keeps no file of its own", key)
		}
		title := one.Title
		if title == "" {
			title = one.Key
		}
		write(map[string]string{
			"key":   one.Key,
			"file":  filepath.ToSlash(one.File),
			"title": title,
		})
		return
	}

	keys := make([]string, 0, len(loaded.Singles))
	for _, one := range loaded.Singles {
		keys = append(keys, one.Key)
	}
	fail("no single called %q; the model has: %v", key, keys)
}

func write(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fail("%v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "reference: "+format+"\n", args...)
	os.Exit(1)
}
