//go:build integration

package admin_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/admin"
	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/mail"
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

func TestAnOrganisationIsOpenedWithEverythingInIt(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}
	organisation, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.tenancy.CreateProject(t.Context(), organisation.ID, "shop", "The shop", "bakery.example"); err != nil {
		t.Fatal(err)
	}

	whole, err := h.web.Organisation(t.Context(), boss, "bakery")
	if err != nil {
		t.Fatalf("Organisation() error = %v, want nil", err)
	}
	if whole.Name != "Bakery" || whole.Slug != "bakery" {
		t.Errorf("Organisation() = %q at %q, want Bakery at bakery", whole.Name, whole.Slug)
	}
	if len(whole.People) != 1 || whole.People[0].Email != owner.Email {
		t.Errorf("Organisation() people = %+v, want only %s", whole.People, owner.Email)
	}
	if len(whole.Projects) != 1 || whole.Projects[0].Name != "The shop" {
		t.Errorf("Organisation() sites = %+v, want The shop", whole.Projects)
	}
}

func TestOpeningOneNeedsAdministering(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}

	// A screen that asks for the sites of an empty organisation gets a list
	// with nothing in it, not nothing at all.
	empty, err := h.web.Organisation(t.Context(), boss, "bakery")
	if err != nil {
		t.Fatal(err)
	}
	if empty.Projects == nil || empty.People == nil {
		t.Errorf("Organisation() = %+v, want empty lists rather than none", empty)
	}

	if _, err := h.web.Organisation(t.Context(), owner, "bakery"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Organisation() as its own owner error = %v, want ErrNotAllowed", err)
	}
	if err := h.web.RemoveOrganisation(t.Context(), owner, "bakery"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("RemoveOrganisation() as its own owner error = %v, want ErrNotAllowed", err)
	}
}

func TestOneThatStillHoldsSitesStays(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}
	organisation, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.tenancy.CreateProject(t.Context(), organisation.ID, "shop", "The shop", "bakery.example"); err != nil {
		t.Fatal(err)
	}

	err = h.web.RemoveOrganisation(t.Context(), boss, "bakery")
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("RemoveOrganisation() with a site still in it error = %v, want a refusal", err)
	}
	if !strings.Contains(refused.Reason, "The shop") {
		t.Errorf("the refusal says %q, want it to name The shop", refused.Reason)
	}

	if _, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery"); err != nil {
		t.Errorf("after the refusal the organisation is gone: %v", err)
	}
}

func TestAnEmptyOneIsRemovedAndTheGitRemoteStays(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	remote := filepath.Join(t.TempDir(), "remote.git")
	cfg := gitsync.Config{GitBin: gitBinary(t)}
	if err := gitsync.InitBareForDevelopment(t.Context(), cfg, remote); err != nil {
		t.Fatal(err)
	}

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}
	organisation, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	project, err := h.tenancy.CreateProject(t.Context(), organisation.ID, "shop", "The shop", "bakery.example")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.db.Queries.SetProjectSource(t.Context(), store.SetProjectSourceParams{
		ID:            pgUUID(project.ID),
		Remote:        remote,
		SourceKind:    "repository",
		SourceAddress: remote,
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.tenancy.RemoveProject(t.Context(), project.ID); err != nil {
		t.Fatal(err)
	}
	if err := h.web.RemoveOrganisation(t.Context(), boss, "bakery"); err != nil {
		t.Fatalf("RemoveOrganisation() error = %v, want nil", err)
	}

	if _, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery"); err == nil {
		t.Error("the organisation is still there after it was removed")
	}
	if _, err := h.web.Organisation(t.Context(), boss, "bakery"); err == nil {
		t.Error("Organisation() still answers for one that was removed")
	}

	// What is in git is nobody's to throw away from here.
	if _, err := os.Stat(filepath.Join(remote, "HEAD")); err != nil {
		t.Errorf("the git remote was touched: %v", err)
	}

	seen, err := h.web.Organisations(t.Context(), boss)
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 0 {
		t.Errorf("%d organisations left, want 0", len(seen))
	}
}

func TestRemovingAnOrganisationIsRecorded(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}
	if err := h.web.RemoveOrganisation(t.Context(), boss, "bakery"); err != nil {
		t.Fatal(err)
	}

	records, err := h.web.Records(t.Context(), boss, 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range records {
		if record.To == "Bakery" && record.Did == string(admin.RemovedAnOrganisation) {
			found = true
		}
	}
	if !found {
		t.Errorf("removing is not in the history: %+v", records)
	}
}

func gitBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("CHATCMS_GIT_BIN"); bin != "" {
		return bin
	}
	found, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}
	return found
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

type postbox struct {
	sent []string
}

func (p *postbox) Post(_ context.Context, to string, _ mail.Message) error {
	p.sent = append(p.sent, to)
	return nil
}

func TestEverybodyIsListedWithWhereTheyBelong(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}
	organisation, err := h.tenancy.OrganisationBySlug(t.Context(), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	project, err := h.tenancy.CreateProject(t.Context(), organisation.ID, "shop", "The shop", "bakery.example")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.tenancy.SetProjectRole(t.Context(), project.ID, mustParse(t, editor.ID), policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	listed, err := h.web.People(t.Context(), boss, "", 0)
	if err != nil {
		t.Fatalf("People() error = %v, want nil", err)
	}
	if listed.Altogether != 3 || int(listed.Shown) != 3 {
		t.Errorf("People() shows %d of %d, want 3 of 3", listed.Shown, listed.Altogether)
	}

	by := map[string]httpapi.PersonWholeView{}
	for _, person := range listed.People {
		by[person.Email] = person
	}

	if places := by[owner.Email].Organisations; len(places) != 1 || places[0].Slug != "bakery" ||
		places[0].Role != string(policy.OrganisationOwner) {
		t.Errorf("the owner belongs to %+v, want owner of bakery", places)
	}
	if sites := by[editor.Email].Projects; len(sites) != 1 || sites[0].Slug != "shop" ||
		sites[0].Organisation != "bakery" || sites[0].Role != string(policy.RoleEditor) {
		t.Errorf("the editor reaches %+v, want editor on bakery/shop", sites)
	}
	if places := by[boss.Email].Organisations; len(places) != 0 {
		t.Errorf("the administrator belongs to %+v, want nowhere", places)
	}
	if !by[boss.Email].Administers {
		t.Error("the administrator is not marked as one")
	}
	if len(by[boss.Email].Projects) != 0 {
		t.Errorf("the administrator reaches %+v, want nothing", by[boss.Email].Projects)
	}
}

func TestSomebodyIsFoundByNameOrAddress(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	h.person(t, "marieke@korenaar.example")
	h.person(t, "sam@korenaar.example")

	found, err := h.web.People(t.Context(), boss, "marieke", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(found.People) != 1 || found.People[0].Email != "marieke@korenaar.example" {
		t.Errorf("searching for marieke found %+v", found.People)
	}

	none, err := h.web.People(t.Context(), boss, "nobody-is-called-this", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(none.People) != 0 {
		t.Errorf("searching for nobody found %+v", none.People)
	}
}

func TestOnlyAnAdministratorReadsThePeople(t *testing.T) {
	h := newHarness(t)
	h.administrator(t, "boss@example.org")
	ordinary := h.person(t, "ordinary@example.org")

	if _, err := h.web.People(t.Context(), ordinary, "", 0); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("People() as an ordinary person error = %v, want ErrNotAllowed", err)
	}
	if _, err := h.web.Person(t.Context(), ordinary, "ordinary@example.org"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Person() about themselves error = %v, want ErrNotAllowed", err)
	}
	if _, err := h.web.MakePerson(t.Context(), ordinary, "new@example.org", ""); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("MakePerson() as an ordinary person error = %v, want ErrNotAllowed", err)
	}
}

func TestOnePersonIsOpenedAndTheirAccountTurnedOff(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	owner := h.person(t, "owner@example.org")

	if _, err := h.web.MakeOrganisation(t.Context(), boss, "bakery", "Bakery", owner.Email); err != nil {
		t.Fatal(err)
	}

	opened, err := h.web.Person(t.Context(), boss, owner.Email)
	if err != nil {
		t.Fatalf("Person() error = %v, want nil", err)
	}
	if !opened.AccountOn || len(opened.Organisations) != 1 {
		t.Errorf("Person() = %+v, want an account that is on, in one organisation", opened)
	}
	if opened.LastSeen != nil {
		t.Errorf("Person() says they were last here at %v, and they never were", opened.LastSeen)
	}

	if err := h.web.TurnAccountOff(t.Context(), boss, owner.Email); err != nil {
		t.Fatal(err)
	}

	again, err := h.web.Person(t.Context(), boss, owner.Email)
	if err != nil {
		t.Fatal(err)
	}
	if again.AccountOn {
		t.Error("the account is still on after it was turned off")
	}
}

func TestOpeningAnAddressNobodyHasSaysSo(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")

	_, err := h.web.Person(t.Context(), boss, "nobody@example.org")
	if err == nil {
		t.Fatal("Person() for an address nobody has error = nil, want a refusal")
	}
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("Person() error = %v, want a refusal with a reason", err)
	}
}

func TestAPersonIsMadeAndSentTheWayIn(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")

	post := &postbox{}
	h.identity.SendsWith(post)

	made, err := h.web.MakePerson(t.Context(), boss, "Sam@Example.org", "Sam Dekker")
	if err != nil {
		t.Fatalf("MakePerson() error = %v, want nil", err)
	}
	if made.Email != "sam@example.org" || made.DisplayName != "Sam Dekker" {
		t.Errorf("MakePerson() = %+v, want sam@example.org as Sam Dekker", made)
	}
	if !made.StillToAccept {
		t.Error("a person made this way has a password already")
	}
	if len(post.sent) != 1 || post.sent[0] != "sam@example.org" {
		t.Errorf("the way in was sent to %v", post.sent)
	}

	// No password was set, so signing in with anything at all is refused.
	if _, _, _, err := h.identity.SignIn(t.Context(), "sam@example.org",
		"a long enough passphrase", "a-caller", false); err == nil {
		t.Error("somebody made by an administrator can be signed in as")
	}

	records, err := h.web.Records(t.Context(), boss, 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range records {
		if record.Did == string(admin.MadeAPerson) && record.To == "sam@example.org" {
			found = true
		}
	}
	if !found {
		t.Errorf("making a person is not in the history: %+v", records)
	}
}

func TestAnAddressThatIsAlreadySomebodyIsRefused(t *testing.T) {
	h := newHarness(t)
	boss := h.administrator(t, "boss@example.org")
	h.person(t, "sam@example.org")

	_, err := h.web.MakePerson(t.Context(), boss, "SAM@example.org", "Somebody Else")
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("MakePerson() at a taken address error = %v, want a refusal", err)
	}
	if !strings.Contains(refused.Reason, "sam@example.org") {
		t.Errorf("the refusal says %q, want it to name the address", refused.Reason)
	}

	// The account that was there is untouched.
	kept, err := h.web.Person(t.Context(), boss, "sam@example.org")
	if err != nil {
		t.Fatal(err)
	}
	if kept.DisplayName == "Somebody Else" {
		t.Error("the account that was already there was renamed")
	}
}
