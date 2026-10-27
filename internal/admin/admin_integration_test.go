//go:build integration

package admin_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/admin"
	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

type harness struct {
	db       *store.DB
	identity *auth.Service
	people   *auth.People
	tenancy  *tenancy.Store
	store    *admin.Store
	web      *admin.Web
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}

	identity := auth.NewService(db, slog.New(slog.NewTextHandler(io.Discard, nil)), "https://chatcms.example")
	tenancyStore := tenancy.NewStore(db)
	adminStore := admin.NewStore(db)

	return &harness{
		db:       db,
		identity: identity,
		people:   identity.People(),
		tenancy:  tenancyStore,
		store:    adminStore,
		web:      admin.NewWeb(admin.NewService(adminStore, tenancyStore, identity)),
	}
}

func (h *harness) person(t *testing.T, email string) httpapi.Actor {
	t.Helper()
	made, err := h.people.Create(t.Context(), email, email, "a long enough passphrase")
	if err != nil {
		t.Fatalf("making %s failed: %v", email, err)
	}
	return httpapi.Actor{ID: made.ID.String(), Email: made.Email, DisplayName: made.DisplayName}
}

func (h *harness) administrator(t *testing.T, email string) httpapi.Actor {
	t.Helper()
	who := h.person(t, email)
	id, err := uuid.Parse(who.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.store.MakeAdministrator(t.Context(), id); err != nil {
		t.Fatalf("appointing %s failed: %v", email, err)
	}
	return who
}

func TestOnlyAnAdministratorMakesAnOrganisation(t *testing.T) {
	h := newHarness(t)
	ordinary := h.person(t, "ordinary@example.org")

	_, err := h.web.MakeOrganisation(t.Context(), ordinary, "bakery", "Bakery", ordinary.Email)
	if !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Fatalf("MakeOrganisation() as an ordinary person error = %v, want ErrNotAllowed", err)
	}

	boss := h.administrator(t, "boss@example.org")
	made, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", ordinary.Email)
	if err != nil {
		t.Fatalf("MakeOrganisation() as an administrator error = %v, want nil", err)
	}
	if made.Slug != "bakery" {
		t.Errorf("slug = %q, want bakery", made.Slug)
	}
}

func TestMakingAnOrganisationDoesNotPutTheAdministratorInIt(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	customer := h.person(t, "customer@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", customer.Email); err != nil {
		t.Fatal(err)
	}

	// The owner named is the customer, and the administrator holds nothing.
	customerActor := h.actorFor(t, customer)
	bossActor := h.actorFor(t, boss)
	organisation := h.organisation(t, "bakery")

	if role, held := customerActor.Organisations[organisation]; !held || role != policy.OrganisationOwner {
		t.Errorf("the named owner holds %q, %v; want owner", role, held)
	}
	if _, held := bossActor.Organisations[organisation]; held {
		t.Error("making an organisation put the administrator into it")
	}
}

func TestAdministeringGrantsNothingOverContent(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	customer := h.person(t, "customer@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", customer.Email); err != nil {
		t.Fatal(err)
	}
	organisation := h.organisation(t, "bakery")
	organisationID, err := uuid.Parse(organisation)
	if err != nil {
		t.Fatal(err)
	}
	project, err := h.tenancy.CreateProject(t.Context(), organisationID, "site", "Site", "site.example")
	if err != nil {
		t.Fatal(err)
	}

	bossActor := h.actorFor(t, boss)
	resource := policy.Project(organisation, project.ID.String())

	for _, action := range []policy.Action{
		policy.ActionReadContent,
		policy.ActionEditContent,
		policy.ActionPublish,
		policy.ActionGoBackToAVersion,
		policy.ActionSeeTechnicalDetail,
	} {
		if policy.Can(bossActor, action, resource) {
			t.Errorf("administering granted %s on a customer's project", action)
		}
	}
}

func TestAnAdministratorSeesEveryOrganisationAndOthersDoNot(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	customer := h.person(t, "customer@example.org")

	for _, slug := range []string{"bakery", "studio"} {
		if _, err := h.web.MakeOrganisation(t.Context(), boss, slug, slug, customer.Email); err != nil {
			t.Fatal(err)
		}
	}

	seen, err := h.web.Organisations(t.Context(), boss)
	if err != nil {
		t.Fatalf("Organisations() error = %v, want nil", err)
	}
	if len(seen) != 2 {
		t.Errorf("an administrator sees %d organisations, want 2", len(seen))
	}

	if _, err := h.web.Organisations(t.Context(), customer); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Organisations() as the customer error = %v, want ErrNotAllowed", err)
	}
}

func TestTheLastAdministratorStays(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")

	err := h.web.StandDown(t.Context(), boss, boss.Email)
	if err == nil {
		t.Fatal("StandDown() as the last administrator error = nil, want a refusal")
	}
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("StandDown() error = %v, want a refusal with a reason", err)
	}

	// With a second one, standing down is allowed.
	second := h.person(t, "second@example.org")
	if err := h.web.Appoint(t.Context(), boss, second.Email); err != nil {
		t.Fatalf("Appoint() error = %v, want nil", err)
	}
	if err := h.web.StandDown(t.Context(), boss, boss.Email); err != nil {
		t.Fatalf("StandDown() with another administrator error = %v, want nil", err)
	}

	left, err := h.store.CountAdministrators(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if left != 1 {
		t.Errorf("%d administrators left, want 1", left)
	}
}

func TestAppointingNeedsAdministering(t *testing.T) {
	h := newHarness(t)
	h.administrator(t, "boss@example.org")
	ordinary := h.person(t, "ordinary@example.org")

	if err := h.web.Appoint(t.Context(), ordinary, ordinary.Email); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Appoint() as an ordinary person error = %v, want ErrNotAllowed", err)
	}
}

func TestWhatWasDoneIsRecorded(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	customer := h.person(t, "customer@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", customer.Email); err != nil {
		t.Fatal(err)
	}

	records, err := h.web.Records(t.Context(), boss, 10)
	if err != nil {
		t.Fatalf("Records() error = %v, want nil", err)
	}
	if len(records) != 1 {
		t.Fatalf("%d records, want 1", len(records))
	}
	if records[0].Did != string(admin.MadeAnOrganisation) {
		t.Errorf("did = %q, want %q", records[0].Did, admin.MadeAnOrganisation)
	}
	if records[0].Who != boss.DisplayName {
		t.Errorf("who = %q, want %q", records[0].Who, boss.DisplayName)
	}

	if _, err := h.web.Records(t.Context(), customer, 10); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Records() as the customer error = %v, want ErrNotAllowed", err)
	}
}

func TestTheFirstAdministratorIsAppointedOnlyOnce(t *testing.T) {
	h := newHarness(t)
	h.person(t, "first@example.org")

	changed, err := h.store.MakeAdministratorByEmail(t.Context(), "first@example.org")
	if err != nil || !changed {
		t.Fatalf("first appointment = %v, %v; want true, nil", changed, err)
	}

	changed, err = h.store.MakeAdministratorByEmail(t.Context(), "first@example.org")
	if err != nil || changed {
		t.Fatalf("second appointment = %v, %v; want false, nil", changed, err)
	}

	if _, err := h.store.MakeAdministratorByEmail(t.Context(), "nobody@example.org"); !errors.Is(err, admin.ErrNoSuchPerson) {
		t.Errorf("appointing an address nobody holds error = %v, want ErrNoSuchPerson", err)
	}
}

func (h *harness) actorFor(t *testing.T, who httpapi.Actor) policy.Actor {
	t.Helper()
	id, err := uuid.Parse(who.ID)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := h.tenancy.ActorFor(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	return actor
}

func (h *harness) organisation(t *testing.T, slug string) string {
	t.Helper()
	found, err := h.tenancy.OrganisationBySlug(t.Context(), slug)
	if err != nil {
		t.Fatal(err)
	}
	return found.ID.String()
}

func TestAnAccountTurnedOffCannotSignInAndItsSessionsStop(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	leaving := h.person(t, "leaving@example.org")

	// They are here and signed in before anything happens.
	signedIn, _, _, err := h.identity.SignIn(t.Context(), "leaving@example.org",
		"a long enough passphrase", "a-caller", false)
	if err != nil {
		t.Fatalf("signing in before error = %v, want nil", err)
	}
	if signedIn.Email != "leaving@example.org" {
		t.Fatalf("signed in as %q", signedIn.Email)
	}

	if err := h.web.TurnAccountOff(t.Context(), boss, "leaving@example.org"); err != nil {
		t.Fatalf("TurnAccountOff() error = %v, want nil", err)
	}

	if _, _, _, err := h.identity.SignIn(t.Context(), "leaving@example.org",
		"a long enough passphrase", "a-caller", false); err == nil {
		t.Error("an account that is off signed in")
	}

	// An account that is off whose session keeps working is not really off.
	live, err := h.identity.Sessions().LiveCount(t.Context(), mustParse(t, leaving.ID))
	if err != nil {
		t.Fatal(err)
	}
	if live != 0 {
		t.Errorf("%d sessions still live, want none", live)
	}

	// Turning it on again gives back the password it had.
	if err := h.web.TurnAccountOn(t.Context(), boss, "leaving@example.org"); err != nil {
		t.Fatalf("TurnAccountOn() error = %v, want nil", err)
	}
	if _, _, _, err := h.identity.SignIn(t.Context(), "leaving@example.org",
		"a long enough passphrase", "a-caller", false); err != nil {
		t.Errorf("signing in after it was turned on again error = %v, want nil", err)
	}
}

func TestNobodyTurnsOffTheirOwnAccount(t *testing.T) {
	// It cannot be undone by the person who did it.
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")

	err := h.web.TurnAccountOff(t.Context(), boss, "boss@example.org")

	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("TurnAccountOff() on your own error = %v, want a refusal", err)
	}
}

func TestAnOrganisationOwnerCannotTurnAnAccountOff(t *testing.T) {
	// An account belongs to the installation, not to an organisation:
	// somebody in two would otherwise be locked out of one by the other.
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")
	other := h.person(t, "other@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}

	if err := h.web.TurnAccountOff(t.Context(), owner, other.Email); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("TurnAccountOff() as an organisation owner error = %v, want ErrNotAllowed", err)
	}
}

func mustParse(t *testing.T, id string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
