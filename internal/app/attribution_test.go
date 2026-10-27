package app

import (
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/httpapi"
)

func TestAVersionWrittenWhileActingNamesBoth(t *testing.T) {
	marieke := auth.Person{
		ID:          uuid.New(),
		Email:       "marieke@korenaar.example",
		DisplayName: "Marieke de Vries",
	}

	acting := httpapi.WithActor(t.Context(), httpapi.Actor{
		ID:          marieke.ID.String(),
		Email:       marieke.Email,
		DisplayName: marieke.DisplayName,
		Really: &httpapi.Who{
			ID:          uuid.New().String(),
			Email:       "pim@example.org",
			DisplayName: "Pim Snel",
		},
	})

	wrote := whoWrote(acting, marieke)
	if wrote.Name != "Pim Snel as Marieke de Vries" {
		t.Errorf("the author reads %q, want Pim Snel as Marieke de Vries", wrote.Name)
	}
	// The address is the administrator's, because that is who was really at
	// the keyboard; reading the name is then not the only way to tell.
	if wrote.Email != "pim@example.org" {
		t.Errorf("the author's address is %q, want the administrator's", wrote.Email)
	}
}

func TestAVersionWrittenByThePersonThemselvesNamesOnlyThem(t *testing.T) {
	marieke := auth.Person{
		ID:          uuid.New(),
		Email:       "marieke@korenaar.example",
		DisplayName: "Marieke de Vries",
	}

	herself := httpapi.WithActor(t.Context(), httpapi.Actor{
		ID:          marieke.ID.String(),
		Email:       marieke.Email,
		DisplayName: marieke.DisplayName,
	})

	wrote := whoWrote(herself, marieke)
	if wrote.Name != "Marieke de Vries" || wrote.Email != marieke.Email {
		t.Errorf("the author is %+v, want her alone", wrote)
	}
}

func TestAVersionWrittenWithNoRequestIsThePersonAlone(t *testing.T) {
	// A background job has no request and nobody acting; it must still be
	// able to attribute a version.
	sam := auth.Person{ID: uuid.New(), Email: "sam@korenaar.example", DisplayName: "Sam"}

	wrote := whoWrote(t.Context(), sam)
	if wrote.Name != "Sam" || wrote.Email != sam.Email {
		t.Errorf("the author is %+v, want Sam alone", wrote)
	}
}

func TestActingAsSomebodyElseDoesNotRenameAThirdPerson(t *testing.T) {
	// The actor in the request is one person; the version being attributed
	// is another's. Nothing is combined, because the two are not the same.
	sam := auth.Person{ID: uuid.New(), Email: "sam@korenaar.example", DisplayName: "Sam"}

	acting := httpapi.WithActor(t.Context(), httpapi.Actor{
		ID:          uuid.New().String(),
		Email:       "marieke@korenaar.example",
		DisplayName: "Marieke de Vries",
		Really:      &httpapi.Who{Email: "pim@example.org", DisplayName: "Pim Snel"},
	})

	wrote := whoWrote(acting, sam)
	if wrote.Name != "Sam" {
		t.Errorf("the author reads %q, want Sam", wrote.Name)
	}
}
