// Command seed puts one organisation, two projects and three people into a
// database, and gives one project a real site with a remote of its own on
// disk, so that every screen can be looked at with something on it. It is a
// development tool and is not part of the server.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: seed <database url>")
		os.Exit(2)
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, err := store.Open(ctx, os.Args[1])
	if err != nil {
		fail(err)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		fail(err)
	}

	identity := auth.NewService(db, logger, "http://127.0.0.1:18090")
	tenancyStore := tenancy.NewStore(db)
	service := tenancy.NewService(tenancyStore, identity)

	owner, err := identity.People().Create(ctx, "owner@example.org", "Joris Bakker", "a long enough passphrase")
	if err != nil {
		fail(err)
	}

	organisation, err := tenancyStore.CreateOrganisation(ctx, "technative", "TechNative", owner.ID)
	if err != nil {
		fail(err)
	}

	for _, site := range []struct {
		slug, name, domain string
		withASite          bool
	}{
		{"korenaar", "Korenaar Bakery", "korenaar.example", true},
		{"studio-noord", "Studio Noord", "studionoord.example", false},
	} {
		project, err := tenancyStore.CreateProject(ctx, organisation.ID, site.slug, site.name, site.domain)
		if err != nil {
			fail(err)
		}
		if err := tenancyStore.SetProjectRole(ctx, project.ID, owner.ID, policy.RoleOwner); err != nil {
			fail(err)
		}
		if site.withASite {
			if err := bringASiteIn(ctx, db, project.ID, owner); err != nil {
				fail(err)
			}
		}
	}

	who := httpActor(owner)
	for _, invited := range []struct {
		email string
		role  policy.Role
	}{
		{"marieke@korenaar.example", policy.RoleEditor},
		{"sam@korenaar.example", policy.RoleViewer},
	} {
		if _, err := service.InvitePerson(ctx, who, "technative", "korenaar",
			invited.email, string(invited.role)); err != nil {
			fail(err)
		}
	}

	fmt.Println("seeded: owner@example.org with a long enough passphrase")
	fmt.Println("korenaar has a site with a remote under " + stateDir() + "/dev-remotes")
	fmt.Println("studio-noord has none, so the onboarding screen has something to do")
}

func stateDir() string {
	if named := os.Getenv("CHATCMS_STATE_DIR"); named != "" {
		return named
	}
	return "./.chatcms-state"
}

func gitBin() string {
	if named := os.Getenv("CHATCMS_GIT_BIN"); named != "" {
		return named
	}
	return "git"
}

// bringASiteIn gives a project a working copy and a remote on disk, so that
// publishing, the history and going back all have something real to act on
// without anybody having to own a git host.
func bringASiteIn(ctx context.Context, db *store.DB, projectID uuid.UUID, owner auth.Person) error {
	cfg := gitsync.Config{
		GitBin: gitBin(),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@localhost"},
	}

	sites := gitsync.NewSites(cfg, filepath.Join(stateDir(), "sites"), nil)
	author := gitsync.Identity{Name: owner.DisplayName, Email: owner.Email}

	// A real site, when one is named, so that the same acceptance tests can be
	// run against a customer's own repository rather than against the fixture.
	if given := os.Getenv("CHATCMS_SEED_REMOTE"); given != "" {
		repository, err := sites.Clone(ctx, projectID.String(), given)
		if err != nil {
			return err
		}
		defer repository.Close()

		_, err = db.Pool().Exec(ctx,
			`UPDATE projects SET remote = $2, source_kind = 'repository', source_address = $2 WHERE id = $1`,
			pgtype.UUID{Bytes: projectID, Valid: true}, given)
		return err
	}

	// An absolute path, because git runs from the working copy's own
	// directory and a relative remote would be looked for there.
	remote, err := filepath.Abs(filepath.Join(stateDir(), "dev-remotes", projectID.String()+".git"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(remote), 0o700); err != nil {
		return err
	}
	if err := gitsync.InitBareForDevelopment(ctx, cfg, remote); err != nil {
		return err
	}

	repository, err := sites.Create(ctx, projectID.String(), remote, aSmallSite(), author, "Bring the site in")
	if err != nil {
		return err
	}
	defer repository.Close()

	if _, err := repository.Publish(ctx); err != nil {
		return err
	}

	// A little unpublished work, so the publish screen has something to show
	// and the history has something waiting in it.
	if _, err := repository.Save(ctx, gitsync.Save{
		Message: "Edit the home page",
		Author:  gitsync.Identity{Name: "Marieke de Vries", Email: "marieke@korenaar.example"},
		Writes: map[string][]byte{
			"content/_index.md":             []byte(editedHomePage),
			"content/news/spring-prices.md": []byte("---\ntitle: Spring prices\ndate: 2026-03-01T09:00:00Z\n---\n\nOur prices for spring are in.\n"),
		},
	}); err != nil {
		return err
	}

	_, err = db.Pool().Exec(ctx,
		`UPDATE projects SET remote = $2, source_kind = 'repository', source_address = $2 WHERE id = $1`,
		pgtype.UUID{Bytes: projectID, Valid: true}, remote)
	return err
}

// publishedHomePage is the page as the live site has it, and editedHomePage
// is the same page with one word changed, so that the review before
// publishing has a word to show as having changed.
const publishedHomePage = `---
title: Korenaar Bakery
intro: We bake with apples and pears every morning.
openToday: true
style: traditional
accent: '#8b5a2b'
specialities:
  - Sourdough
  - Apple tart
opening:
  - day: Monday
    hours: 7 to 18
  - day: Saturday
    hours: 7 to 16
seo:
  description: A bakery on the Oude Gracht.
---

Come in and smell it.
`

const editedHomePage = `---
title: Korenaar Bakery
intro: We bake with plums and pears every morning.
openToday: true
style: traditional
accent: '#8b5a2b'
specialities:
  - Sourdough
  - Apple tart
opening:
  - day: Monday
    hours: 7 to 18
  - day: Saturday
    hours: 7 to 16
seo:
  description: A bakery on the Oude Gracht.
---

Come in and smell it.
`

// aLoaf is a picture, so that the media library has something in it and the
// field that picks one has something to pick.
func aLoaf() []byte {
	raw, err := base64.StdEncoding.DecodeString(loafPNG)
	if err != nil {
		panic(err)
	}
	return raw
}

const loafPNG = "" +
	"iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAlElEQVR4nO3PMQ3DUADE0EAsnCAppsIqgdyQyd+S" +
	"pRs93Lu+9+dxv7HT+uu0QwHoQwHoQwHoQ68BlqOrD0D3Aeg+AN37AZajqw9A9wHoPgDd+wGWo6sPQPcB6D4A3fsB" +
	"lqOrD0D3Aeg+AN37AZajqw9A9wHoPgDd+wGWo6sPQPcB6D4A3fsBlqOrD0D3Aeg+AN3rAX9D+3FaNCiw0QAAAABJ" +
	"RU5ErkJggg=="

func aSmallSite() fs.FS {
	return fstest.MapFS{
		"hugo.toml": {Data: []byte("baseURL = \"https://korenaar.example/\"\nlanguageCode = \"nl\"\ntitle = \"Korenaar Bakery\"\n")},
		"quiqr/model/base.yaml": {Data: []byte(`singles:
  - key: home
    title: Home page
    file: content/_index.md
    fields:
      - key: title
        type: string
        title: Page title
      - key: intro
        type: string
        title: Introduction
        multiLine: true
      - key: openToday
        type: boolean
        title: Open today
      - key: style
        type: select
        title: Shop style
        options:
          - value: traditional
            text: Traditional
          - value: modern
            text: Modern
      - key: accent
        type: color
        title: Accent colour
      - key: specialities
        type: leaf-array
        title: Specialities
        field:
          key: speciality
          type: string
      - key: opening
        type: accordion
        title: Opening hours
        fields:
          - key: day
            type: string
            title: Day
          - key: hours
            type: string
            title: Hours
      - key: seo
        type: section
        title: Search
        fields:
          - key: description
            type: string
            title: Description
      - key: mainContent
        type: markdown
        title: Page text
collections:
  - key: news
    title: News
    folder: content/news
    itemtitle: article
    hideIndex: true
    fields:
      - key: title
        type: string
        title: Headline
      - key: date
        type: date
        title: Date
      - key: mainContent
        type: markdown
        title: Page text
`)},
		"static/images/loaf.png":       {Data: aLoaf()},
		"content/_index.md":            {Data: []byte(publishedHomePage)},
		"content/news/_index.md":       {Data: []byte("---\ntitle: News\n---\n")},
		"layouts/index.html":           {Data: []byte("<!doctype html>\n<html lang=\"nl\"><head><meta charset=\"utf-8\"><title>{{ .Site.Title }}</title></head>\n<body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n")},
		"layouts/_default/single.html": {Data: []byte("<!doctype html>\n<html lang=\"nl\"><head><meta charset=\"utf-8\"><title>{{ .Title }}</title></head>\n<body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n")},
		"layouts/_default/list.html":   {Data: []byte("<!doctype html>\n<html lang=\"nl\"><head><meta charset=\"utf-8\"><title>{{ .Title }}</title></head>\n<body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n")},
	}
}

func httpActor(person auth.Person) httpapi.Actor {
	return httpapi.Actor{ID: person.ID.String(), Email: person.Email, DisplayName: person.DisplayName}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "seed:", err)
	os.Exit(1)
}
