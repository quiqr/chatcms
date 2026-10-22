// Command seed puts one organisation, one project and three people into a
// database, so that a screen can be looked at with something on it. It is a
// development tool and is not part of the server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/quiqr/chatcms/internal/auth"
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

	for _, site := range []struct{ slug, name, domain string }{
		{"korenaar", "Korenaar Bakery", "korenaar.example"},
		{"studio-noord", "Studio Noord", "studionoord.example"},
	} {
		project, err := tenancyStore.CreateProject(ctx, organisation.ID, site.slug, site.name, site.domain)
		if err != nil {
			fail(err)
		}
		if err := tenancyStore.SetProjectRole(ctx, project.ID, owner.ID, policy.RoleOwner); err != nil {
			fail(err)
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
}

func httpActor(person auth.Person) httpapi.Actor {
	return httpapi.Actor{ID: person.ID.String(), Email: person.Email, DisplayName: person.DisplayName}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "seed:", err)
	os.Exit(1)
}
