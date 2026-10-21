package model

import "io/fs"

// Load reads a site's Quiqr model from fsys, which is rooted at the site
// directory. It reads only, makes no network request, and returns ErrNoModel
// when the site carries no model at all.
func Load(fsys fs.FS) (*Model, error) {
	assembly, err := Assemble(fsys)
	if err != nil {
		return nil, err
	}
	return Decode(assembly)
}
