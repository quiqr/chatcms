//go:build integration

package tenancy_test

import (
	"errors"
	"testing"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

func TestANewOrganisationHasNoLimits(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)

	usage, err := h.store.UsageFor(t.Context(), organisation.ID)
	if err != nil {
		t.Fatalf("UsageFor() error = %v, want nil", err)
	}
	if usage.Members.Limited() || usage.Projects.Limited() {
		t.Errorf("a new organisation has limits: %+v", usage)
	}
	if usage.Members.Held != 1 {
		t.Errorf("it holds %d members, want its admin", usage.Members.Held)
	}
	if usage.Members.Full() || usage.Projects.Full() {
		t.Error("an organisation with no limits is full")
	}
}

func TestASeatIsHeldByAnInvitationNobodyHasAccepted(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)

	two := 2
	if err := h.store.SetLimits(t.Context(), organisation.ID, &two, nil); err != nil {
		t.Fatal(err)
	}

	// Somebody invited and not yet accepted has been taken on, so they hold
	// a seat; otherwise the limit could be walked past by inviting everybody
	// and waiting.
	if _, err := h.inviter.InviteToOrganisation(t.Context(), organisation.ID,
		"waiting@example.org", policy.OrganisationMember); err != nil {
		t.Fatalf("inviting into the second seat failed: %v", err)
	}

	usage, err := h.store.UsageFor(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Members.Held != 2 || !usage.Members.Full() {
		t.Fatalf("after the second invitation the organisation holds %+v", usage.Members)
	}

	_, err = h.inviter.InviteToOrganisation(t.Context(), organisation.ID,
		"third@example.org", policy.OrganisationMember)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("inviting a third error = %v, want a refusal", err)
	}
	if refused.Key != "refused.no-room-for-a-member" {
		t.Errorf("the refusal is named %q", refused.Key)
	}
	// The numbers are what makes it something somebody can act on.
	if refused.Detail["would"] != 3 || refused.Detail["allowed"] != 2 {
		t.Errorf("the refusal says %+v, want 3 and 2", refused.Detail)
	}
}

func TestChangingTheRoleOfSomebodyAlreadyInTakesNoNewSeat(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	other := h.person(t, "other@example.org")

	if err := h.store.SetOrganisationRole(t.Context(), organisation.ID, other.ID,
		policy.OrganisationMember); err != nil {
		t.Fatal(err)
	}

	two := 2
	if err := h.store.SetLimits(t.Context(), organisation.ID, &two, nil); err != nil {
		t.Fatal(err)
	}

	// A full organisation must still be one somebody can reorganise.
	if err := h.store.PutSomebodyInWithRoom(t.Context(), organisation.ID, other.ID,
		policy.OrganisationAdmin); err != nil {
		t.Errorf("changing a role in a full organisation error = %v, want nil", err)
	}
}

func TestASiteIsRefusedWhenThereIsNoRoomForIt(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	h.project(t, organisation, "korenaar")

	one := 1
	if err := h.store.SetLimits(t.Context(), organisation.ID, nil, &one); err != nil {
		t.Fatal(err)
	}

	_, err := h.service.CreateProject(t.Context(), h.who(admin), "technative",
		"second", "A second site", "")
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("making a second site error = %v, want a refusal", err)
	}
	if refused.Key != "refused.no-room-for-a-site" {
		t.Errorf("the refusal is named %q", refused.Key)
	}

	// And nothing was made on the way to refusing.
	projects, err := h.store.ProjectsIn(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("the organisation holds %d sites after the refusal", len(projects))
	}
}

func TestALimitBelowWhatIsHeldRemovesNothing(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	h.project(t, organisation, "korenaar")
	h.project(t, organisation, "studio")

	// Written down after the fact, which is how an agreement is usually
	// recorded: what is there stays, and nothing new is allowed.
	one := 1
	if err := h.store.SetLimits(t.Context(), organisation.ID, nil, &one); err != nil {
		t.Fatalf("lowering the limit below what is held error = %v, want nil", err)
	}

	projects, err := h.store.ProjectsIn(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("lowering the limit removed something: %d sites left", len(projects))
	}

	usage, err := h.store.UsageFor(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !usage.Projects.Full() || usage.Projects.Left() != 0 {
		t.Errorf("an organisation over its limit is not full: %+v", usage.Projects)
	}

	if _, err := h.service.CreateProject(t.Context(), h.who(admin), "technative",
		"third", "A third site", ""); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("making a site over the limit error = %v, want a refusal", err)
	}
}

func TestALimitIsANumberOfAtLeastOne(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)

	none := 0
	err := h.store.SetLimits(t.Context(), organisation.ID, &none, nil)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("a limit of zero error = %v, want a refusal", err)
	}
}
