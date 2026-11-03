package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// AdministrationService is what somebody who administers the installation may
// do. Every method refuses anybody else, which the handlers do not check for
// themselves.
type AdministrationService interface {
	MakeOrganisation(ctx context.Context, who Actor, slug, name, ownerEmail string) (OrganisationView, error)
	PutSomebodyInto(ctx context.Context, who Actor, slug, email, role string) error
	TakeSomebodyOut(ctx context.Context, who Actor, slug, email string) error
	Appoint(ctx context.Context, who Actor, email string) error
	StandDown(ctx context.Context, who Actor, email string) error
	Records(ctx context.Context, who Actor, most int32) ([]RecordView, error)
	Administrators(ctx context.Context, who Actor) ([]AdministratorView, error)
	Organisations(ctx context.Context, who Actor, asked OrganisationsAsked) ([]OrganisationSummaryView, error)
	// MayAdminister answers the one question the interface needs in order to
	// know whether to offer any of this, without making it guess from a
	// refusal.
	TurnAccountOff(ctx context.Context, who Actor, email string) error
	TurnAccountOn(ctx context.Context, who Actor, email string) error
	Organisation(ctx context.Context, who Actor, slug string) (OrganisationWholeView, error)
	RemoveOrganisation(ctx context.Context, who Actor, slug string) error
	SetLimits(ctx context.Context, who Actor, slug string, members, projects *int) error
	People(ctx context.Context, who Actor, search string, most int32) (PeopleView, error)
	Person(ctx context.Context, who Actor, email string) (PersonWholeView, error)
	MakePerson(ctx context.Context, who Actor, email, displayName string) (PersonWholeView, error)
	ActAs(ctx context.Context, who Actor, token, email string) error
	StopActingAs(ctx context.Context, who Actor, token string) error
	MayAdminister(ctx context.Context, who Actor) bool
	HowMuchThereIs(ctx context.Context, who Actor) (organisations, users int64, err error)
}

// RecordView is one thing an administrator did.
type RecordView struct {
	Who    string         `json:"who"`
	Did    string         `json:"did"`
	To     string         `json:"to"`
	Detail map[string]any `json:"detail,omitempty"`
	At     time.Time      `json:"at"`
}

// AdministratorView is somebody who administers the installation.
type AdministratorView struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Yourself    bool   `json:"yourself"`
}

// OrganisationWholeView is one organisation with everything about it, which
// is what an administrator opens.
type OrganisationWholeView struct {
	Slug     string        `json:"slug"`
	Name     string        `json:"name"`
	People   []MemberView  `json:"people"`
	Projects []ProjectView `json:"projects"`
	Usage    UsageView     `json:"usage"`
}

// BelongingView is one place somebody can reach, and what they hold there.
type BelongingView struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	// Organisation is the address of the organisation a site is in, and is
	// empty for an organisation itself.
	Organisation string `json:"organisation,omitempty"`
	Role         string `json:"role"`
}

// PersonWholeView is one person and everywhere they can reach.
type PersonWholeView struct {
	PersonID      string          `json:"personId"`
	Email         string          `json:"email"`
	DisplayName   string          `json:"displayName"`
	Administers   bool            `json:"administers"`
	AccountOn     bool            `json:"accountOn"`
	StillToAccept bool            `json:"stillToAccept"`
	LastSeen      *time.Time      `json:"lastSeen,omitempty"`
	Organisations []BelongingView `json:"organisations"`
	Projects      []BelongingView `json:"projects"`
}

// PeopleView is a page of them, and how many there are altogether, so that
// the screen can say what it is not showing.
type PeopleView struct {
	People     []PersonWholeView `json:"people"`
	Shown      int64             `json:"shown"`
	Altogether int64             `json:"altogether"`
}

// OrganisationSummaryView is one organisation as an administrator sees it,
// which includes ones they are not in.
type OrganisationSummaryView struct {
	Slug     string    `json:"slug"`
	Name     string    `json:"name"`
	Projects int64     `json:"projects"`
	People   int64     `json:"people"`
	Usage    UsageView `json:"usage"`
	// Admins is who administers it, which is empty for one nobody does: the
	// screen says so in the colour it gives a warning.
	Admins []WhoView  `json:"admins"`
	MadeBy string     `json:"madeBy,omitempty"`
	MadeOn *time.Time `json:"madeOn,omitempty"`
}

// WhoView is somebody named in passing, with no more about them than the
// screen shows.
type WhoView struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// OrganisationsAsked is what the list was asked for. The server answers it,
// because a thousand organisations are not a thing to send to a browser to
// be counted there.
type OrganisationsAsked struct {
	Search      string
	AtALimit    bool
	WithNoAdmin bool
}

// CountedView is one of the two things an organisation can run out of: what
// it holds, and what it is allowed. Allowed is absent where there is no
// limit, which is what a new organisation has.
type CountedView struct {
	Held    int  `json:"held"`
	Allowed *int `json:"allowed,omitempty"`
	Full    bool `json:"full"`
}

// UsageView is how full an organisation is. Every screen that says so reads
// this, so that two of them cannot disagree.
type UsageView struct {
	Members  CountedView `json:"members"`
	Projects CountedView `json:"projects"`
}

type limitsRequest struct {
	Members  *int `json:"members"`
	Projects *int `json:"projects"`
}

type makeOrganisationRequest struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	OwnerEmail string `json:"ownerEmail"`
}

type organisationPersonRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type administratorRequest struct {
	Email string `json:"email"`
}

type makePersonRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func administrationHandlers(mux *http.ServeMux, service AdministrationService) {
	mux.Handle("GET /api/administration/me", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			who := ActorFrom(r.Context())
			if !service.MayAdminister(r.Context(), who) {
				WriteJSON(w, http.StatusOK, map[string]any{"administers": false})
				return
			}

			// The counts the area's shell shows beside its nav, so that the
			// shell is not two requests of its own on every navigation.
			organisations, users, err := service.HowMuchThereIs(r.Context(), who)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{
				"administers":   true,
				"organisations": organisations,
				"users":         users,
			})
		})))

	mux.Handle("GET /api/administration/organisations", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			asked := OrganisationsAsked{
				Search:      r.URL.Query().Get("search"),
				AtALimit:    r.URL.Query().Get("filter") == "at-a-limit",
				WithNoAdmin: r.URL.Query().Get("filter") == "no-admin",
			}
			organisations, err := service.Organisations(r.Context(), ActorFrom(r.Context()), asked)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"organisations": organisations})
		})))

	mux.Handle("POST /api/administration/organisations", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body makeOrganisationRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			made, err := service.MakeOrganisation(r.Context(), ActorFrom(r.Context()), body.Slug, body.Name, body.OwnerEmail)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusCreated, made)
		})))

	mux.HandleFunc("/api/administration/organisations", MethodNotAllowed)

	mux.Handle("POST /api/administration/organisations/{organisation}/people", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body organisationPersonRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.PutSomebodyInto(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), body.Email, body.Role); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("DELETE /api/administration/organisations/{organisation}/people/{email}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.TakeSomebodyOut(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET /api/administration/organisations/{organisation}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			whole, err := service.Organisation(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, whole)
		})))

	mux.Handle("DELETE /api/administration/organisations/{organisation}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.RemoveOrganisation(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("PUT /api/administration/organisations/{organisation}/limits", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body limitsRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.SetLimits(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), body.Members, body.Projects); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc("/api/administration/organisations/{organisation}/limits", MethodNotAllowed)

	mux.HandleFunc("/api/administration/organisations/{organisation}", MethodNotAllowed)

	mux.Handle("GET /api/administration/administrators", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			people, err := service.Administrators(r.Context(), ActorFrom(r.Context()))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"administrators": people})
		})))

	mux.Handle("POST /api/administration/administrators", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body administratorRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.Appoint(r.Context(), ActorFrom(r.Context()), body.Email); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc("/api/administration/administrators", MethodNotAllowed)

	mux.Handle("DELETE /api/administration/administrators/{email}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.StandDown(r.Context(), ActorFrom(r.Context()), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("POST /api/administration/accounts/{email}/off", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.TurnAccountOff(r.Context(), ActorFrom(r.Context()), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("POST /api/administration/accounts/{email}/on", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.TurnAccountOn(r.Context(), ActorFrom(r.Context()), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET /api/administration/people", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			people, err := service.People(r.Context(), ActorFrom(r.Context()),
				r.URL.Query().Get("search"), howMany(r.URL.Query().Get("most")))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, people)
		})))

	mux.Handle("POST /api/administration/people", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body makePersonRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			made, err := service.MakePerson(r.Context(), ActorFrom(r.Context()),
				body.Email, body.DisplayName)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusCreated, made)
		})))

	mux.HandleFunc("/api/administration/people", MethodNotAllowed)

	mux.Handle("GET /api/administration/people/{email}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			person, err := service.Person(r.Context(), ActorFrom(r.Context()), r.PathValue("email"))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, person)
		})))

	mux.HandleFunc("/api/administration/people/{email}", MethodNotAllowed)

	mux.Handle("POST /api/administration/people/{email}/act-as", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.ActAs(r.Context(), ActorFrom(r.Context()),
				sessionValue(r), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	// Ending is not under /administration, because whoever is acting is not
	// an administrator as far as every other answer is concerned: the way
	// out has to be reachable by the person they are acting as.
	mux.Handle("DELETE /api/acting-as", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.StopActingAs(r.Context(), ActorFrom(r.Context()),
				sessionValue(r)); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET /api/administration/records", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			records, err := service.Records(r.Context(), ActorFrom(r.Context()), 100)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"records": records})
		})))
}

// howMany reads the asked-for size of a page. Anything unreadable is the
// default rather than an error: a screen asking for nonsense gets a page.
func howMany(asked string) int32 {
	if asked == "" {
		return 0
	}
	most, err := strconv.Atoi(asked)
	if err != nil {
		return 0
	}
	return int32(most)
}

// sessionValue is the value the session cookie carries, which is how acting
// as somebody is put on this session rather than on every session the
// administrator has.
func sessionValue(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}
