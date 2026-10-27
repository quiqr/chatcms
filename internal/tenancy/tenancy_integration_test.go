//go:build integration

package tenancy_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

const password = "a long enough passphrase"

type harness struct {
	store   *tenancy.Store
	service *tenancy.Service
	auth    *auth.Service
	logs    *bytes.Buffer
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

	logs := &bytes.Buffer{}
	identity := auth.NewService(db, slog.New(slog.NewJSONHandler(logs, nil)), "https://chatcms.example")
	tenancyStore := tenancy.NewStore(db)

	return &harness{
		store:   tenancyStore,
		service: tenancy.NewService(tenancyStore, identity),
		auth:    identity,
		logs:    logs,
	}
}

func (h *harness) person(t *testing.T, email string) auth.Person {
	t.Helper()

	person, err := h.auth.People().Create(t.Context(), email, email, password)
	if err != nil {
		t.Fatalf("creating a person failed: %v", err)
	}
	return person
}

func (h *harness) who(person auth.Person) httpapi.Actor {
	return httpapi.Actor{ID: person.ID.String(), Email: person.Email}
}

func (h *harness) organisation(t *testing.T, slug string, owner auth.Person) tenancy.Organisation {
	t.Helper()

	organisation, err := h.store.CreateOrganisation(t.Context(), slug, slug, owner.ID)
	if err != nil {
		t.Fatalf("creating the organisation failed: %v", err)
	}
	return organisation
}

func (h *harness) project(t *testing.T, organisation tenancy.Organisation, slug string) tenancy.Project {
	t.Helper()

	project, err := h.store.CreateProject(t.Context(), organisation.ID, slug, slug, slug+".example")
	if err != nil {
		t.Fatalf("creating the project failed: %v", err)
	}
	return project
}

func TestCreatingAnOrganisationMakesItsCreatorAnOwner(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")

	organisation := h.organisation(t, "technative", owner)

	actor, err := h.store.ActorFor(t.Context(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if actor.Organisations[organisation.ID.String()] != policy.OrganisationOwner {
		t.Errorf("the creator holds %q, want owner", actor.Organisations[organisation.ID.String()])
	}
}

func TestSlugsAreCheckedAndUnique(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")

	for _, slug := range []string{"", "With Capitals", "trailing-", "double--hyphen", "a_b"} {
		if _, err := h.store.CreateOrganisation(t.Context(), slug, "A name", owner.ID); !errors.Is(err, tenancy.ErrUnusableSlug) {
			t.Errorf("CreateOrganisation(%q) error = %v, want ErrUnusableSlug", slug, err)
		}
	}

	h.organisation(t, "technative", owner)
	if _, err := h.store.CreateOrganisation(t.Context(), "technative", "Again", owner.ID); !errors.Is(err, tenancy.ErrSlugTaken) {
		t.Errorf("a second organisation with the same address gave %v, want ErrSlugTaken", err)
	}
}

func TestAProjectAddressIsUniqueOnlyInItsOrganisation(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")

	first := h.organisation(t, "technative", owner)
	second := h.organisation(t, "another", owner)

	h.project(t, first, "korenaar")
	if _, err := h.store.CreateProject(t.Context(), second.ID, "korenaar", "Korenaar", ""); err != nil {
		t.Errorf("the same project address in another organisation was refused: %v", err)
	}
	if _, err := h.store.CreateProject(t.Context(), first.ID, "korenaar", "Korenaar", ""); !errors.Is(err, tenancy.ErrSlugTaken) {
		t.Errorf("a repeated address in one organisation gave %v, want ErrSlugTaken", err)
	}
}

func TestMembershipsLoadForAPerson(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")

	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	actor, err := h.store.ActorFor(t.Context(), editor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if actor.Projects[project.ID.String()] != policy.RoleEditor {
		t.Errorf("the editor holds %q", actor.Projects[project.ID.String()])
	}
	if len(actor.Organisations) != 0 {
		t.Errorf("the editor is in %d organisations, want none", len(actor.Organisations))
	}
}

func TestListingShowsOnlyWhatIsReachable(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	organisation := h.organisation(t, "technative", owner)
	reachable := h.project(t, organisation, "korenaar")
	h.project(t, organisation, "studio-noord")

	if err := h.store.SetProjectRole(t.Context(), reachable.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	seenByOwner, err := h.service.ListProjects(t.Context(), h.who(owner), "technative")
	if err != nil {
		t.Fatal(err)
	}
	if len(seenByOwner) != 2 {
		t.Errorf("the owner sees %d projects, want 2", len(seenByOwner))
	}

	seenByEditor, err := h.service.ListProjects(t.Context(), h.who(editor), "technative")
	if err != nil {
		t.Fatal(err)
	}
	if len(seenByEditor) != 1 || seenByEditor[0].Slug != "korenaar" {
		t.Errorf("the editor sees %+v, want only the project they hold a role on", seenByEditor)
	}
}

func TestSomebodyElsesProjectLooksMissing(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	stranger := h.person(t, "stranger@example.org")

	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	_, theirs := h.service.GetProject(t.Context(), h.who(stranger), "technative", "korenaar")
	_, absent := h.service.GetProject(t.Context(), h.who(stranger), "technative", "no-such-project")

	if !errors.Is(theirs, httpapi.ErrMissing) {
		t.Errorf("somebody else's project gave %v, want ErrMissing", theirs)
	}
	if !errors.Is(absent, httpapi.ErrMissing) {
		t.Errorf("a project that is not there gave %v, want ErrMissing", absent)
	}
}

func TestAnEditorMayNotManagePeople(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")
	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.ListPeople(t.Context(), h.who(editor), "technative", "korenaar"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("an editor listing people gave %v, want ErrNotAllowed", err)
	}
	if _, err := h.service.InvitePerson(t.Context(), h.who(editor), "technative", "korenaar", "new@example.org", "viewer"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("an editor inviting gave %v, want ErrNotAllowed", err)
	}
}

func TestInvitingSomeoneNew(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	member, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"customer@example.org", "editor")
	if err != nil {
		t.Fatalf("InvitePerson() error = %v, want nil", err)
	}
	if member.Email != "customer@example.org" || member.Role != "editor" {
		t.Errorf("member = %+v", member)
	}
	if !member.StillToAccept {
		t.Error("a new person is not marked as still to accept")
	}

	person, err := h.auth.People().ByEmail(t.Context(), "customer@example.org")
	if err != nil {
		t.Fatalf("the invited person was not created: %v", err)
	}
	if person.HasPassword {
		t.Error("the invited person already has a password")
	}
	if !bytes.Contains(h.logs.Bytes(), []byte("set-password?token=")) {
		t.Error("no link that sets a first password was issued")
	}
}

func TestInvitingSomeoneKnownMakesNoSecondPerson(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	known := h.person(t, "known@example.org")

	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	member, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"KNOWN@example.org", "viewer")
	if err != nil {
		t.Fatalf("InvitePerson() error = %v, want nil", err)
	}
	if member.PersonID != known.ID.String() {
		t.Errorf("a second person was made: %s", member.PersonID)
	}
	if member.StillToAccept {
		t.Error("someone who already has a password is marked as still to accept")
	}
}

func TestInvitingTwiceChangesTheRole(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")

	for _, role := range []string{"editor", "viewer"} {
		if _, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
			"customer@example.org", role); err != nil {
			t.Fatal(err)
		}
	}

	members, err := h.store.Members(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, member := range members {
		if member.Email == "customer@example.org" {
			count++
			if member.Role != policy.RoleViewer {
				t.Errorf("role = %q, want the one invited with second", member.Role)
			}
		}
	}
	if count != 1 {
		t.Errorf("the person appears %d times, want once", count)
	}
}

func TestInvitingWithSomethingThatIsNotARole(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	_, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"customer@example.org", "administrator")
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("InvitePerson() with a made-up role = %v, want a refusal", err)
	}
}

func TestChangingARoleAndRemovingAPerson(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")

	if err := h.store.SetProjectRole(t.Context(), project.ID, owner.ID, policy.RoleOwner); err != nil {
		t.Fatal(err)
	}

	invited, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"customer@example.org", "editor")
	if err != nil {
		t.Fatal(err)
	}

	changed, err := h.service.ChangeRole(t.Context(), h.who(owner), "technative", "korenaar",
		invited.PersonID, "viewer")
	if err != nil {
		t.Fatalf("ChangeRole() error = %v, want nil", err)
	}
	if changed.Role != "viewer" {
		t.Errorf("role = %q, want viewer", changed.Role)
	}

	if err := h.service.RemovePerson(t.Context(), h.who(owner), "technative", "korenaar",
		invited.PersonID); err != nil {
		t.Fatalf("RemovePerson() error = %v, want nil", err)
	}

	members, _ := h.store.Members(t.Context(), project.ID)
	for _, member := range members {
		if member.Email == "customer@example.org" {
			t.Error("the person is still on the project")
		}
	}
}

func TestLastAdministratorCannotBeRemoved(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")

	if err := h.store.SetProjectRole(t.Context(), project.ID, owner.ID, policy.RoleOwner); err != nil {
		t.Fatal(err)
	}

	err := h.service.RemovePerson(t.Context(), h.who(owner), "technative", "korenaar", owner.ID.String())
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("removing the last administrator gave %v, want a refusal", err)
	}
}

func TestLastOrganisationOwnerCannotBeRemoved(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)

	if err := h.store.RemoveFromOrganisation(t.Context(), organisation.ID, owner.ID); !errors.Is(err, tenancy.ErrWouldLeaveNobodyInCharge) {
		t.Errorf("removing the last owner gave %v, want a refusal", err)
	}
}

func TestOrganisationsAreSeparate(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	other := h.person(t, "other@example.org")

	first := h.organisation(t, "technative", owner)
	second := h.organisation(t, "another", other)

	h.project(t, first, "korenaar")
	h.project(t, second, "secret")

	seen, err := h.service.ListProjects(t.Context(), h.who(owner), "technative")
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 {
		t.Errorf("the owner of one organisation sees %d projects, want 1", len(seen))
	}

	if _, err := h.service.ListProjects(t.Context(), h.who(owner), "another"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("listing another organisation's projects gave %v, want ErrMissing", err)
	}
}

func TestProjectViewCarriesWhatThePersonMayDo(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")
	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	view, err := h.service.GetProject(t.Context(), h.who(editor), "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}

	if view.YourRole != "editor" {
		t.Errorf("yourRole = %q", view.YourRole)
	}

	may := map[string]bool{}
	for _, action := range view.YouMay {
		may[action] = true
	}
	if !may[string(policy.ActionEditContent)] {
		t.Error("an editor is not told they may edit content")
	}
	if may[string(policy.ActionManagePeople)] {
		t.Error("an editor is told they may manage people")
	}
	if may[string(policy.ActionSeeTechnicalDetail)] {
		t.Error("an editor is told they may see technical detail")
	}
}

func TestNobodyReachesNothing(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	nobody := httpapi.Anonymous()

	if _, err := h.service.GetProject(t.Context(), nobody, "technative", "korenaar"); err == nil {
		t.Error("nobody reached a project")
	}
	if _, err := h.service.ListOrganisations(t.Context(), nobody); err == nil {
		t.Error("nobody listed organisations")
	}
}

func TestOrganisationMembershipCanBeGivenAndTaken(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	second := h.person(t, "second@example.org")

	organisation := h.organisation(t, "technative", owner)

	if err := h.store.SetOrganisationRole(t.Context(), organisation.ID, second.ID, policy.OrganisationOwner); err != nil {
		t.Fatal(err)
	}

	actor, err := h.store.ActorFor(t.Context(), second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if actor.Organisations[organisation.ID.String()] != policy.OrganisationOwner {
		t.Errorf("the second person holds %q", actor.Organisations[organisation.ID.String()])
	}

	// With two owners, one may be removed.
	if err := h.store.RemoveFromOrganisation(t.Context(), organisation.ID, second.ID); err != nil {
		t.Errorf("removing one of two owners gave %v, want nil", err)
	}
}

func TestOrganisationsForAPersonIncludesOnesReachedThroughAProject(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	editor := h.person(t, "editor@example.org")

	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")
	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	reachable, err := h.store.OrganisationsFor(t.Context(), editor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reachable) != 1 || reachable[0].Slug != "technative" {
		t.Errorf("the editor reaches %+v, want the organisation their project is in", reachable)
	}

	views, err := h.service.ListOrganisations(t.Context(), h.who(editor))
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Errorf("ListOrganisations() = %+v, want one", views)
	}
}

func TestAStrangerReachesNoOrganisation(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	stranger := h.person(t, "stranger@example.org")

	h.organisation(t, "technative", owner)

	views, err := h.service.ListOrganisations(t.Context(), h.who(stranger))
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 0 {
		t.Errorf("a stranger reaches %+v, want nothing", views)
	}
}

func TestCreatingAProjectThroughTheService(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	stranger := h.person(t, "stranger@example.org")

	h.organisation(t, "technative", owner)

	view, err := h.service.CreateProject(t.Context(), h.who(owner), "technative",
		"korenaar", "Korenaar Bakery", "korenaar.example")
	if err != nil {
		t.Fatalf("CreateProject() error = %v, want nil", err)
	}
	if view.Slug != "korenaar" || view.YourRole != "owner" {
		t.Errorf("view = %+v", view)
	}

	if _, err := h.service.CreateProject(t.Context(), h.who(stranger), "technative",
		"another", "Another", ""); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("a stranger creating a project gave %v, want ErrMissing", err)
	}

	if _, err := h.service.CreateProject(t.Context(), h.who(owner), "technative",
		"Not A Slug", "Another", ""); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("an unusable address gave %v, want a refusal", err)
	}
}

func TestChangingToSomethingThatIsNotARole(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	project := h.project(t, organisation, "korenaar")

	if err := h.store.SetProjectRole(t.Context(), project.ID, owner.ID, policy.RoleOwner); err != nil {
		t.Fatal(err)
	}

	member, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"customer@example.org", "editor")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.ChangeRole(t.Context(), h.who(owner), "technative", "korenaar",
		member.PersonID, "administrator"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("changing to a made-up role gave %v, want a refusal", err)
	}

	if _, err := h.service.ChangeRole(t.Context(), h.who(owner), "technative", "korenaar",
		"not-an-identifier", "viewer"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("changing the role of nobody gave %v, want ErrMissing", err)
	}

	if err := h.service.RemovePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"not-an-identifier"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("removing nobody gave %v, want ErrMissing", err)
	}
}

func TestAnOrganisationOwnerReachesAProjectTheyWereNeverAddedTo(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	view, err := h.service.GetProject(t.Context(), h.who(owner), "technative", "korenaar")
	if err != nil {
		t.Fatalf("GetProject() error = %v, want nil", err)
	}
	if view.YourRole != "owner" {
		t.Errorf("yourRole = %q, want owner through the organisation", view.YourRole)
	}
}

func TestInvitingWithNoAddress(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "technative", owner)
	h.project(t, organisation, "korenaar")

	if _, err := h.service.InvitePerson(t.Context(), h.who(owner), "technative", "korenaar",
		"   ", "editor"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("inviting nobody gave %v, want a refusal", err)
	}
}

func TestAProjectCanBeChangedAfterItIsSetUp(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "bakery", owner)
	h.project(t, organisation, "korenaar")

	changed, err := h.service.ChangeProject(t.Context(), h.who(owner), "bakery", "korenaar",
		"Korenaar BV", "korenaar.nl")
	if err != nil {
		t.Fatalf("ChangeProject() error = %v, want nil", err)
	}

	if changed.Name != "Korenaar BV" || changed.Domain != "korenaar.nl" {
		t.Errorf("changed = %+v, want the new name and address", changed)
	}
	// The address it is reached at is what links already sent are written
	// against, so it does not move with the name.
	if changed.Slug != "korenaar" {
		t.Errorf("slug = %q, want it unchanged", changed.Slug)
	}
}

func TestAProjectNeedsAName(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "bakery", owner)
	h.project(t, organisation, "korenaar")

	_, err := h.service.ChangeProject(t.Context(), h.who(owner), "bakery", "korenaar", "   ", "")

	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("ChangeProject() with an empty name error = %v, want a refusal", err)
	}
}

func TestOnlySomebodyAllowedChangesAProject(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "bakery", owner)
	project := h.project(t, organisation, "korenaar")

	editor := h.person(t, "editor@example.org")
	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.ChangeProject(t.Context(), h.who(editor), "bakery", "korenaar", "Mine", ""); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("ChangeProject() as an editor error = %v, want ErrNotAllowed", err)
	}

	if err := h.service.RemoveProject(t.Context(), h.who(editor), "bakery", "korenaar"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("RemoveProject() as an editor error = %v, want ErrNotAllowed", err)
	}
}

func TestRemovingAProjectForgetsItAndNothingElse(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "bakery", owner)
	h.project(t, organisation, "korenaar")
	h.project(t, organisation, "studio")

	if err := h.service.RemoveProject(t.Context(), h.who(owner), "bakery", "korenaar"); err != nil {
		t.Fatalf("RemoveProject() error = %v, want nil", err)
	}

	left, err := h.service.ListProjects(t.Context(), h.who(owner), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 1 || left[0].Slug != "studio" {
		t.Errorf("what is left = %+v, want only studio", left)
	}

	// Asking for the one that is gone says there is nothing there, rather
	// than failing in some other way.
	if _, err := h.service.ChangeProject(t.Context(), h.who(owner), "bakery", "korenaar", "Back", ""); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("changing a removed project error = %v, want ErrMissing", err)
	}
}

func TestAnOrganisationOwnerInvitesToIt(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	h.organisation(t, "bakery", owner)

	invited, err := h.service.InviteToOrganisation(t.Context(), h.who(owner), "bakery",
		"marieke@korenaar.example", string(policy.OrganisationMember))
	if err != nil {
		t.Fatalf("InviteToOrganisation() error = %v, want nil", err)
	}

	// An address nobody held gets an account and a way in.
	if !invited.StillToAccept {
		t.Error("somebody with no account was not sent a way in")
	}

	people, err := h.service.OrganisationPeople(t.Context(), h.who(owner), "bakery")
	if err != nil {
		t.Fatal(err)
	}
	if len(people) != 2 {
		t.Errorf("%d people in the organisation, want the owner and the invited", len(people))
	}
}

func TestInvitingSomebodyWhoAlreadyHasAnAccountSendsNothing(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	h.person(t, "marieke@example.org")
	h.organisation(t, "bakery", owner)

	invited, err := h.service.InviteToOrganisation(t.Context(), h.who(owner), "bakery",
		"marieke@example.org", string(policy.OrganisationMember))
	if err != nil {
		t.Fatalf("InviteToOrganisation() error = %v, want nil", err)
	}
	if invited.StillToAccept {
		t.Error("somebody who already has a password was sent a way in they do not need")
	}
}

func TestOnlyAnOwnerManagesAnOrganisationsPeople(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	organisation := h.organisation(t, "bakery", owner)

	member := h.person(t, "member@example.org")
	if err := h.store.SetOrganisationRole(t.Context(), organisation.ID, member.ID, policy.OrganisationMember); err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.InviteToOrganisation(t.Context(), h.who(member), "bakery",
		"somebody@example.org", string(policy.OrganisationMember)); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("inviting as a member error = %v, want ErrNotAllowed", err)
	}

	if err := h.service.ChangeOrganisationRole(t.Context(), h.who(member), "bakery",
		owner.ID.String(), string(policy.OrganisationMember)); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("changing a role as a member error = %v, want ErrNotAllowed", err)
	}

	if err := h.service.RemoveFromOrganisation(t.Context(), h.who(member), "bakery",
		owner.ID.String()); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("removing as a member error = %v, want ErrNotAllowed", err)
	}
}

func TestTheLastOwnerOfAnOrganisationStays(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	h.organisation(t, "bakery", owner)

	err := h.service.RemoveFromOrganisation(t.Context(), h.who(owner), "bakery", owner.ID.String())

	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("removing the last owner error = %v, want a refusal with a reason", err)
	}
}

func TestAnInvitationIsSentAgainAndTheOldOneStops(t *testing.T) {
	h := newHarness(t)
	owner := h.person(t, "owner@example.org")
	h.organisation(t, "bakery", owner)

	invited, err := h.service.InviteToOrganisation(t.Context(), h.who(owner), "bakery",
		"marieke@korenaar.example", string(policy.OrganisationMember))
	if err != nil {
		t.Fatal(err)
	}

	if err := h.service.SendTheInvitationAgain(t.Context(), h.who(owner), "bakery", invited.PersonID); err != nil {
		t.Fatalf("SendTheInvitationAgain() error = %v, want nil", err)
	}

	// Somebody who has accepted has nothing to send again.
	if err := h.service.SendTheInvitationAgain(t.Context(), h.who(owner), "bakery", owner.ID.String()); err == nil {
		t.Error("sending again to somebody who has accepted was allowed")
	}
}
