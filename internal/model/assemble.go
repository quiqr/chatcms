package model

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	modelDir      = "quiqr/model"
	includesDir   = modelDir + "/includes"
	partialsDir   = modelDir + "/partials"
	remoteCache   = modelDir + "/partialsRemoteCache"
	legacyBase    = "sukoh"
	mergePartial  = "_mergePartial"
	dogfoodPrefix = "dogfood_site://"
	filePrefix    = "file://"
)

var listKeys = []string{"collections", "singles", "dynamics", "menu"}

// Assembly is the generic tree of a site's model together with where it came
// from and anything worth telling the implementer.
type Assembly struct {
	BaseFile string
	Tree     map[string]any
	Warnings []Warning
}

// Assemble reads a site's model files and merges them in the order the Quiqr
// desktop app uses. It reads only from fsys and writes nothing.
func Assemble(fsys fs.FS) (*Assembly, error) {
	base, ok := locateBase(fsys)
	if !ok {
		return nil, ErrNoModel
	}

	tree, err := readFile(fsys, base)
	if err != nil {
		return nil, err
	}

	assembly := &Assembly{BaseFile: base}
	assembly.Tree = assembly.skeleton(tree)

	if err := assembly.mergeIncludes(fsys); err != nil {
		return nil, err
	}
	if err := assembly.appendSubIncludes(fsys); err != nil {
		return nil, err
	}
	if err := assembly.mergePartials(fsys); err != nil {
		return nil, err
	}

	assembly.dropEmptyLists()
	return assembly, nil
}

func locateBase(fsys fs.FS) (string, bool) {
	for _, ext := range Extensions {
		candidate := modelDir + "/base." + ext
		if exists(fsys, candidate) {
			return candidate, true
		}
	}
	for _, ext := range Extensions {
		candidate := legacyBase + "." + ext
		if exists(fsys, candidate) {
			return candidate, true
		}
	}
	return "", false
}

func exists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func readFile(fsys fs.FS, name string) (map[string]any, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, &Error{File: name, Want: "a readable model file", Err: err}
	}
	return Parse(name, data)
}

// skeleton makes sure the four list keys exist as lists, which is what the
// reference application does before merging anything into them. A key that is
// present but is not a list is replaced, as it is there, and reported, which
// it is not.
func (a *Assembly) skeleton(tree map[string]any) map[string]any {
	out := clone(tree).(map[string]any)
	for _, key := range listKeys {
		if _, ok := out[key].([]any); ok {
			continue
		}
		if existing, present := out[key]; present && existing != nil {
			a.Warnings = append(a.Warnings, Warning{
				File:    a.BaseFile,
				Path:    key,
				Message: "this should be a list; what is there was ignored",
			})
		}
		out[key] = []any{}
	}
	return out
}

func (a *Assembly) dropEmptyLists() {
	for _, key := range listKeys {
		if list, ok := a.Tree[key].([]any); ok && len(list) == 0 {
			delete(a.Tree, key)
		}
	}
}

func (a *Assembly) mergeIncludes(fsys fs.FS) error {
	files, err := modelFilesIn(fsys, includesDir)
	if err != nil {
		return err
	}

	for _, name := range files {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return &Error{File: name, Want: "a readable include", Err: err}
		}

		tree, err := parseAny(name, data)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(path.Base(name), path.Ext(name))

		if list, ok := tree.([]any); ok && isListKey(key) {
			existing, _ := a.Tree[key].([]any)
			a.Tree[key] = append(append([]any{}, existing...), list...)
			continue
		}

		object, ok := tree.(map[string]any)
		if !ok {
			continue
		}

		existing, hasExisting := a.Tree[key]
		if !hasExisting {
			a.Tree[key] = object
			continue
		}
		a.Tree[key] = Merge(object, existing)
	}
	return nil
}

func (a *Assembly) appendSubIncludes(fsys fs.FS) error {
	subs := []struct {
		dir string
		key string
	}{
		{includesDir + "/singles", "singles"},
		{includesDir + "/collections", "collections"},
		{includesDir + "/menus", "menu"},
	}

	for _, sub := range subs {
		files, err := modelFilesIn(fsys, sub.dir)
		if err != nil {
			return err
		}
		for _, name := range files {
			item, err := readFile(fsys, name)
			if err != nil {
				return err
			}
			existing, _ := a.Tree[sub.key].([]any)
			a.Tree[sub.key] = append(existing, item)
		}
	}
	return nil
}

func (a *Assembly) mergePartials(fsys fs.FS) error {
	for _, key := range []string{"collections", "singles", "dynamics"} {
		list, ok := a.Tree[key].([]any)
		if !ok {
			continue
		}

		merged := make([]any, 0, len(list))
		for index, entry := range list {
			item, ok := entry.(map[string]any)
			if !ok {
				merged = append(merged, entry)
				continue
			}
			resolved, err := a.mergePartialInto(fsys, item, key, index)
			if err != nil {
				return err
			}
			merged = append(merged, resolved)
		}
		a.Tree[key] = merged
	}
	return nil
}

func (a *Assembly) mergePartialInto(fsys fs.FS, item map[string]any, key string, index int) (map[string]any, error) {
	reference, ok := item[mergePartial].(string)
	if !ok || reference == "" {
		return item, nil
	}

	where := pathOf(key, index)

	file, err := a.resolvePartial(fsys, reference, where)
	if err != nil {
		return nil, err
	}
	if file == "" {
		return item, nil
	}

	partial, err := readFile(fsys, file)
	if err != nil {
		return nil, err
	}

	merged := Merge(partial, item).(map[string]any)
	merged["fields"] = dedupeFields(merged["fields"])
	if fields, ok := merged["fields"].([]any); ok && len(fields) == 0 {
		delete(merged, "fields")
	}
	delete(merged, mergePartial)
	return merged, nil
}

// resolvePartial returns the file a _mergePartial reference names. An empty
// name with no error means the reference was reported as a warning and the
// item is used unmerged, which is what the reference application does when a
// partial cannot be found.
func (a *Assembly) resolvePartial(fsys fs.FS, reference, where string) (string, error) {
	switch {
	case strings.HasPrefix(reference, dogfoodPrefix):
		a.Warnings = append(a.Warnings, Warning{
			File: a.BaseFile, Path: where,
			Message: "this site uses " + reference + ", which only the Quiqr desktop app provides",
		})
		return "", nil

	case strings.HasPrefix(reference, "http://"), strings.HasPrefix(reference, "https://"),
		strings.HasPrefix(reference, filePrefix):
		if file, ok := findIn(fsys, remoteCache, encodeReference(reference)); ok {
			return file, nil
		}
		return "", atf(a.BaseFile, where,
			"the partial %q has not been fetched into %s; fetching happens during sync, not while reading the model",
			reference, remoteCache)

	default:
		if file, ok := findIn(fsys, partialsDir, reference); ok {
			return file, nil
		}
		a.Warnings = append(a.Warnings, Warning{
			File: a.BaseFile, Path: where,
			Message: "the partial \"" + reference + "\" is not in " + partialsDir,
		})
		return "", nil
	}
}

// encodeReference mirrors encodeURIComponent, which the desktop app uses to
// name a cached remote partial.
func encodeReference(reference string) string {
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.!~*'()"

	var out strings.Builder
	for _, b := range []byte(reference) {
		if strings.IndexByte(unreserved, b) >= 0 {
			out.WriteByte(b)
			continue
		}
		fmt.Fprintf(&out, "%%%02X", b)
	}
	return out.String()
}

func findIn(fsys fs.FS, dir, name string) (string, bool) {
	if exists(fsys, path.Join(dir, name)) {
		return path.Join(dir, name), true
	}
	for _, ext := range Extensions {
		candidate := path.Join(dir, name+"."+ext)
		if exists(fsys, candidate) {
			return candidate, true
		}
	}
	return "", false
}

// dedupeFields reproduces the reference application's deduplication: reverse,
// keep the first occurrence of each key, reverse again. The last declaration
// of a key wins and the survivors are ordered by where those last occurrences
// fell.
func dedupeFields(value any) any {
	fields, ok := value.([]any)
	if !ok {
		return value
	}

	seen := make(map[string]bool, len(fields))
	kept := make([]any, 0, len(fields))

	for i := len(fields) - 1; i >= 0; i-- {
		field, ok := fields[i].(map[string]any)
		if !ok {
			kept = append(kept, fields[i])
			continue
		}
		key, _ := field["key"].(string)
		if key != "" && seen[key] {
			continue
		}
		if key != "" {
			seen[key] = true
		}
		kept = append(kept, fields[i])
	}

	for left, right := 0, len(kept)-1; left < right; left, right = left+1, right-1 {
		kept[left], kept[right] = kept[right], kept[left]
	}
	return kept
}

func modelFilesIn(fsys fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !supported(extensionOf(entry.Name())) {
			continue
		}
		files = append(files, path.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

func parseAny(name string, data []byte) (any, error) {
	var tree any
	var err error

	switch extensionOf(name) {
	case "json":
		tree, err = parseJSON(data)
	case "toml":
		tree, err = parseTOML(data)
	default:
		tree, err = parseYAML(data)
	}
	if err != nil {
		return nil, &Error{File: name, Want: "a readable model file", Err: err}
	}
	return Normalise(tree), nil
}

func isListKey(key string) bool {
	for _, candidate := range listKeys {
		if candidate == key {
			return true
		}
	}
	return false
}

func pathOf(key string, index int) string {
	return key + "[" + strconv.Itoa(index) + "]"
}
