package onboarding

import (
	"errors"
	"io/fs"
	"sort"
	"strings"

	"github.com/quiqr/chatcms/internal/model"
)

// hugoConfigurations are the names and places Hugo accepts its configuration
// in. A site is taken to be a Hugo site when one of these is there.
var hugoConfigurations = []string{
	"hugo.toml", "hugo.yaml", "hugo.yml", "hugo.json",
	"config.toml", "config.yaml", "config.yml", "config.json",
	"config/_default/hugo.toml", "config/_default/hugo.yaml",
	"config/_default/hugo.yml", "config/_default/hugo.json",
	"config/_default/config.toml", "config/_default/config.yaml",
	"config/_default/config.yml", "config/_default/config.json",
}

// Inspection is what was found in an imported site, in words an implementer
// can act on.
type Inspection struct {
	// IsHugoSite says a Hugo configuration was found.
	IsHugoSite bool `json:"isHugoSite"`
	// HugoConfiguration is the file that said so.
	HugoConfiguration string `json:"hugoConfiguration,omitempty"`
	// HasModel says a Quiqr model was found and could be read.
	HasModel bool `json:"hasModel"`
	// ModelBaseFile is the file the model starts from.
	ModelBaseFile string `json:"modelBaseFile,omitempty"`
	// Pages and Collections are what the model offers, when there is one.
	Pages       int `json:"pages"`
	Collections int `json:"collections"`
	// Problem is what is wrong, in a sentence an implementer can act on. It
	// is empty when nothing is wrong.
	Problem string `json:"problem,omitempty"`
	// LookedFor names what was searched for when something was not found, so
	// that an implementer can see why rather than wonder.
	LookedFor []string `json:"lookedFor,omitempty"`
}

// Inspect says what a site is and what it carries. A site with no model is
// not a failure: it is a site that needs one.
func Inspect(site fs.FS) Inspection {
	inspection := Inspection{}

	for _, candidate := range hugoConfigurations {
		if exists(site, candidate) {
			inspection.IsHugoSite = true
			inspection.HugoConfiguration = candidate
			break
		}
	}

	if !inspection.IsHugoSite {
		inspection.Problem = "This does not look like a Hugo site: none of the files Hugo " +
			"keeps its configuration in is here."
		inspection.LookedFor = append([]string{}, hugoConfigurations...)
		sort.Strings(inspection.LookedFor)
		return inspection
	}

	parsed, err := model.Load(site)
	switch {
	case errors.Is(err, model.ErrNoModel):
		inspection.Problem = "This is a Hugo site with no Quiqr model yet, so there are no " +
			"forms to edit it with. Add a model under quiqr/model to finish setting it up."
		inspection.LookedFor = []string{"quiqr/model/base.yaml", "sukoh.yaml"}
		return inspection

	case err != nil:
		inspection.Problem = "This site has a Quiqr model that cannot be read: " + err.Error()
		return inspection
	}

	inspection.HasModel = true
	inspection.ModelBaseFile = parsed.BaseFile
	inspection.Pages = len(parsed.Singles)
	inspection.Collections = len(parsed.Collections)
	return inspection
}

// Ready says the site can be edited now.
func (i Inspection) Ready() bool {
	return i.IsHugoSite && i.HasModel
}

func exists(site fs.FS, name string) bool {
	info, err := fs.Stat(site, strings.TrimPrefix(name, "./"))
	return err == nil && !info.IsDir()
}
