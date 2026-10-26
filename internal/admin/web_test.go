package admin

import (
	"errors"
	"strings"
	"testing"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

func TestARoleAnOrganisationDoesNotHaveIsRefused(t *testing.T) {
	// Falling back to the smaller role when somebody meant the larger one is
	// a silent demotion, so an unknown word is refused instead.
	web := &Web{service: &Service{}}

	// A project role is not an organisation role, which is the mistake worth
	// catching; it is named through the policy rather than written out, so
	// that this file does not itself carry a role as a value.
	for _, asked := range []string{
		"",
		string(policy.RoleEditor),
		"administrator",
		"superuser",
		strings.ToUpper(string(policy.OrganisationOwner)),
	} {
		err := web.PutSomebodyInto(t.Context(), httpapi.Actor{ID: "someone"}, "bakery", "a@b.c", asked)

		var refused *httpapi.RefusedError
		if !errors.As(err, &refused) {
			t.Errorf("PutSomebodyInto() with role %q error = %v, want a refusal", asked, err)
		}
	}
}

func TestWhatAnActWasCalledReadsAsASentence(t *testing.T) {
	// The record is read by a person, so each one has to finish the sentence
	// "Pim Snel ...".
	for _, did := range []What{
		MadeAnOrganisation,
		RenamedAnOrganisation,
		RemovedAnOrganisation,
		PutSomebodyIn,
		TookSomebodyOut,
		AppointedAnAdmin,
		StoodDownAnAdmin,
	} {
		if did == "" {
			t.Error("an act has no name")
		}
		if first := string(did)[0]; first >= 'A' && first <= 'Z' {
			t.Errorf("%q starts with a capital, so it does not follow a name", did)
		}
	}
}

func TestARefusalKeepsItsReason(t *testing.T) {
	given := &httpapi.RefusedError{Reason: "that address is already taken"}
	if got := refusal(given); !errors.Is(got, given) {
		t.Errorf("refusal() lost the reason: %v", got)
	}

	other := errors.New("the database fell over")
	if got := refusal(other); !errors.Is(got, other) {
		t.Errorf("refusal() changed an unrelated failure into %v", got)
	}
}
