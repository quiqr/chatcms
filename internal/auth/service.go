package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/mail"
	"github.com/quiqr/chatcms/internal/store"
)

// Service joins the parts: it checks a claim, issues a session, and hands out
// the single-use links. It is what the HTTP layer talks to.
type Service struct {
	identity    Identity
	people      *People
	sessions    *Sessions
	credentials *Credentials
	limiter     *Limiter
	logger      *slog.Logger
	baseURL     string
	// post is where a message is handed to be sent. It is nil on an install
	// with no mail server, and the link goes to the log instead.
	post mail.Posting
}

// SendsWith gives the service somewhere to post messages. It is set after
// construction because the queue the post uses is built later than the
// service that signs people in.
func (s *Service) SendsWith(post mail.Posting) { s.post = post }

func NewService(db *store.DB, logger *slog.Logger, baseURL string) *Service {
	return &Service{
		identity:    NewPasswordIdentity(db),
		people:      NewPeople(db),
		sessions:    NewSessions(db),
		credentials: NewCredentials(db),
		limiter:     NewLimiter(),
		logger:      logger,
		baseURL:     strings.TrimRight(baseURL, "/"),
	}
}

// People exposes the people store, which the tests and the tenancy epic use
// to create someone.
func (s *Service) People() *People { return s.people }

// Sessions exposes the session store, which the tests read.
func (s *Service) Sessions() *Sessions { return s.sessions }

// Credentials exposes the link store, which invitations use.
func (s *Service) Credentials() *Credentials { return s.credentials }

// SignIn checks a claim and, if it holds, issues a session.
func (s *Service) SignIn(
	ctx context.Context,
	email, password, caller string,
	remember bool,
) (httpapi.Actor, string, time.Time, error) {
	address := strings.ToLower(strings.TrimSpace(email))

	if err := s.limiter.Allow(address, caller); err != nil {
		return httpapi.Actor{}, "", time.Time{}, httpapi.ErrTooManyAttempts
	}

	person, err := s.identity.Authenticate(ctx, Claim{Email: address, Password: password})
	if err != nil {
		// Every reason is the same reason as far as the caller is concerned.
		if errors.Is(err, ErrNoSuchPerson) || errors.Is(err, ErrWrongPassword) ||
			errors.Is(err, ErrNoPassword) || errors.Is(err, ErrDisabled) {
			s.limiter.RecordFailure(address, caller)
			return httpapi.Actor{}, "", time.Time{}, httpapi.ErrSignInRefused
		}
		return httpapi.Actor{}, "", time.Time{}, err
	}

	token, session, err := s.sessions.Issue(ctx, person.ID, remember)
	if err != nil {
		return httpapi.Actor{}, "", time.Time{}, err
	}

	s.limiter.RecordSuccess(address, caller)
	return actorOf(person), token, session.ExpiresAt, nil
}

// SignOut revokes a session.
func (s *Service) SignOut(ctx context.Context, token string) error {
	return s.sessions.Revoke(ctx, token)
}

// SignOutEverywhere ends every session this account holds, including the one
// asking. Somebody who thinks their account is not theirs any more wants
// every door shut, including the one they are standing in.
func (s *Service) SignOutEverywhere(ctx context.Context, who httpapi.Actor) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}
	return s.sessions.RevokeAll(ctx, personID)
}

// Lookup turns a session value into the person it identifies. A value that
// identifies nobody is not an error.
func (s *Service) Lookup(ctx context.Context, token string) (httpapi.Actor, error) {
	person, session, err := s.sessions.Lookup(ctx, token)
	if errors.Is(err, ErrNoSession) {
		return httpapi.Anonymous(), nil
	}
	if err != nil {
		return httpapi.Anonymous(), err
	}

	if session.ActingAs == nil {
		return actorOf(person), nil
	}

	// While somebody is being acted as, they are who the request is from.
	// Who is really here rides along, and is what the record and the author
	// of a version are written from.
	acted, err := s.people.ByID(ctx, *session.ActingAs)
	if err != nil || acted.Off {
		// Whoever was being acted as is gone, or their account was turned
		// off while this went on. Being themselves again is the safe answer;
		// the alternative is a session that answers nothing.
		return actorOf(person), nil
	}

	actor := actorOf(acted)
	actor.Really = &httpapi.Who{
		ID:          person.ID.String(),
		Email:       person.Email,
		DisplayName: person.DisplayName,
	}
	return actor, nil
}

// ActAs begins acting as somebody else on this session. It says nothing
// about whether that is allowed: admin asks the policy first.
func (s *Service) ActAs(ctx context.Context, token string, personID uuid.UUID) error {
	return s.sessions.ActAs(ctx, token, personID)
}

// StopActingAs ends it. It is never refused: a way out that can be refused
// is a trap.
func (s *Service) StopActingAs(ctx context.Context, token string) error {
	return s.sessions.StopActingAs(ctx, token)
}

// Invite issues the link that sets someone's first password and sends it.
func (s *Service) Invite(ctx context.Context, personID uuid.UUID) error {
	return s.InviteTo(ctx, personID, "", "")
}

// InviteTo is Invite, with the site and the person doing the inviting, which
// is what makes the message read as something a person wrote rather than as
// a notification from a machine.
func (s *Service) InviteTo(ctx context.Context, personID uuid.UUID, siteName, invitedBy string) error {
	token, expires, err := s.credentials.Issue(ctx, KindInvitation, personID)
	if err != nil {
		return err
	}

	person, err := s.people.ByID(ctx, personID)
	if err != nil {
		return err
	}

	s.send(ctx, "invitation", person, token, expires,
		mail.Invitation(siteName, invitedBy, s.linkFor(token)))
	return nil
}

// RequestPasswordReset issues a reset link when the address is known, and
// says nothing either way.
func (s *Service) RequestPasswordReset(ctx context.Context, email, caller string) error {
	address := strings.ToLower(strings.TrimSpace(email))

	if err := s.limiter.Allow(address, caller); err != nil {
		return httpapi.ErrTooManyAttempts
	}
	s.limiter.RecordFailure(address, caller)

	person, err := s.people.ByEmail(ctx, address)
	if errors.Is(err, ErrNoSuchPerson) {
		// Nothing to do, and nothing to say.
		return nil
	}
	if err != nil {
		return err
	}

	token, expires, err := s.credentials.Issue(ctx, KindPasswordReset, person.ID)
	if err != nil {
		return err
	}
	s.send(ctx, "password reset", person, token, expires,
		mail.PasswordReset(s.linkFor(token)))
	return nil
}

// SetPasswordWithToken spends an invitation or a reset link and signs that
// person in. Somebody who has just proved they hold the address and chosen a
// password has done everything signing in would ask for; making them do it
// again is a form to fill in for nothing, and the first thing an invited
// person would meet.
func (s *Service) SetPasswordWithToken(ctx context.Context, token, password string) (httpapi.Actor, string, time.Time, error) {
	person, err := s.credentials.Spend(ctx, token, password, s.people, s.sessions)

	switch {
	case errors.Is(err, ErrBadToken):
		return httpapi.Anonymous(), "", time.Time{}, httpapi.ErrLinkUnusable
	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooCommon):
		return httpapi.Anonymous(), "", time.Time{}, &httpapi.PasswordRejectedError{Reason: err.Error()}
	case err != nil:
		return httpapi.Anonymous(), "", time.Time{}, err
	}

	// Not remembered: a link from an email is often opened on somebody
	// else's machine, and a session that outlives the browser is one they
	// did not ask for.
	sessionToken, session, err := s.sessions.Issue(ctx, person.ID, false)
	if err != nil {
		return httpapi.Anonymous(), "", time.Time{}, err
	}

	s.send(ctx, "password changed", person, "", time.Time{}, mail.PasswordChanged(""))

	return actorOf(person), sessionToken, session.ExpiresAt, nil
}

func (s *Service) linkFor(token string) string {
	return s.baseURL + "/set-password?token=" + url.QueryEscape(token)
}

// confirmLinkFor is where a link that proves a new address lands, which is
// not where a link that sets a password lands.
func (s *Service) confirmLinkFor(token string) string {
	return s.baseURL + "/confirm-address?token=" + url.QueryEscape(token)
}

// send puts the message on its way. When there is nowhere to send, the post
// writes the link to the log instead, which is what an install run by the
// person who owns it needs: being made to set up SMTP before anybody can
// sign in would be a poor welcome.
func (s *Service) send(ctx context.Context, kind string, person Person, token string, expires time.Time, message mail.Message) {
	if s.post == nil {
		s.recordLink(ctx, kind, person.ID, token, expires)
		return
	}

	if err := s.post.Post(ctx, person.Email, message); err != nil {
		// The link exists either way, so it is written down rather than
		// lost, and the failure is said out loud.
		s.logger.LogAttrs(ctx, slog.LevelError,
			"a message could not be put on its way",
			slog.String("kind", kind),
			slog.String("error", err.Error()),
		)
		s.recordLink(ctx, kind, person.ID, token, expires)
	}
}

// recordLink writes the link where an implementer can read it, which is what
// happens when there is nowhere to send mail.
func (s *Service) recordLink(ctx context.Context, kind string, personID uuid.UUID, token string, expires time.Time) {
	s.logger.LogAttrs(ctx, slog.LevelWarn,
		"a sign-in link was issued and could not be sent, because sending mail is not set up",
		slog.String("kind", kind),
		slog.String("person", personID.String()),
		slog.String("link", s.linkFor(token)),
		slog.Time("expires", expires),
	)
}

func actorOf(person Person) httpapi.Actor {
	return httpapi.Actor{
		ID:          person.ID.String(),
		Email:       person.Email,
		DisplayName: person.DisplayName,
	}
}

// TurnAccountOff stops an account being used and ends every session it has.
// An account that is off whose session keeps working until it expires is an
// account that is not really off.
func (s *Service) TurnAccountOff(ctx context.Context, who httpapi.Actor, personID uuid.UUID) error {
	if who.ID == personID.String() {
		return &httpapi.RefusedError{
			Key:    "refused.cannot-turn-off-own-account",
			Reason: "you cannot turn off your own account; ask somebody else to",
		}
	}

	if err := s.people.TurnOff(ctx, personID); err != nil {
		return err
	}
	return s.sessions.RevokeAll(ctx, personID)
}

// MakeSomebody gives an address an account with no password and sends the
// way in. Nobody but the person themselves ever sets the password: one that
// somebody else typed is one nobody can prove was never written down.
func (s *Service) MakeSomebody(ctx context.Context, email, displayName string) (Person, error) {
	address := strings.ToLower(strings.TrimSpace(email))
	if address == "" {
		return Person{}, &httpapi.RefusedError{Key: "refused.email-address-needed", Reason: "an email address is needed"}
	}

	if _, err := s.people.ByEmail(ctx, address); err == nil {
		return Person{}, &httpapi.RefusedError{
			Key:    "refused.address-has-an-account",
			Reason: address + " already has an account",
			Detail: map[string]any{"address": address},
		}
	} else if !errors.Is(err, ErrNoSuchPerson) {
		return Person{}, err
	}

	person, err := s.people.Create(ctx, address, strings.TrimSpace(displayName), "")
	if err != nil {
		return Person{}, err
	}
	if err := s.Invite(ctx, person.ID); err != nil {
		return Person{}, err
	}
	return person, nil
}

// TurnAccountOn lets it be used again, with the password it had.
func (s *Service) TurnAccountOn(ctx context.Context, personID uuid.UUID) error {
	return s.people.TurnOn(ctx, personID)
}

// ChangeOwnName changes what somebody is called. Only their own.
func (s *Service) ChangeOwnName(ctx context.Context, who httpapi.Actor, displayName string) (httpapi.Actor, error) {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.Anonymous(), httpapi.ErrNotAllowed
	}
	if strings.TrimSpace(displayName) == "" {
		return httpapi.Anonymous(), &httpapi.RefusedError{Key: "refused.name-cannot-empty", Reason: "a name cannot be empty"}
	}

	person, err := s.people.Rename(ctx, personID, displayName)
	if err != nil {
		return httpapi.Anonymous(), err
	}
	return actorOf(person), nil
}

// ChangeOwnPassword changes somebody's password, asking for the one they
// have first. A session left open on somebody else's machine should not be
// enough to take an account over.
func (s *Service) ChangeOwnPassword(ctx context.Context, who httpapi.Actor, current, next string) error {
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return httpapi.ErrNotAllowed
	}

	if _, err := s.identity.Authenticate(ctx, Claim{Email: who.Email, Password: current}); err != nil {
		return &httpapi.RefusedError{Key: "refused.not-current-password", Reason: "that is not your current password"}
	}

	if err := s.people.SetPassword(ctx, personID, next); err != nil {
		switch {
		case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooCommon):
			return &httpapi.PasswordRejectedError{Reason: err.Error()}
		}
		return err
	}

	s.send(ctx, "password changed", Person{ID: personID, Email: who.Email}, "", time.Time{},
		mail.PasswordChanged(""))
	return nil
}

// StopEveryInvitationFor makes every unused invitation for somebody useless,
// which is what sending a new one has to do first.
func (s *Service) StopEveryInvitationFor(ctx context.Context, personID uuid.UUID) error {
	return s.credentials.StopEveryLinkFor(ctx, KindInvitation, personID)
}
