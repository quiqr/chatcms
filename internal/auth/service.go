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
}

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

// Lookup turns a session value into the person it identifies. A value that
// identifies nobody is not an error.
func (s *Service) Lookup(ctx context.Context, token string) (httpapi.Actor, error) {
	person, _, err := s.sessions.Lookup(ctx, token)
	if errors.Is(err, ErrNoSession) {
		return httpapi.Anonymous(), nil
	}
	if err != nil {
		return httpapi.Anonymous(), err
	}
	return actorOf(person), nil
}

// Invite issues the link that sets someone's first password.
func (s *Service) Invite(ctx context.Context, personID uuid.UUID) error {
	token, expires, err := s.credentials.Issue(ctx, KindInvitation, personID)
	if err != nil {
		return err
	}
	s.recordLink(ctx, "invitation", personID, token, expires)
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
	s.recordLink(ctx, "password reset", person.ID, token, expires)
	return nil
}

// SetPasswordWithToken spends an invitation or a reset link.
func (s *Service) SetPasswordWithToken(ctx context.Context, token, password string) error {
	_, err := s.credentials.Spend(ctx, token, password, s.people, s.sessions)

	switch {
	case errors.Is(err, ErrBadToken):
		return httpapi.ErrLinkUnusable
	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooCommon):
		return &httpapi.PasswordRejectedError{Reason: err.Error()}
	}
	return err
}

// recordLink writes the link where an implementer can read it. There is no
// way to send mail yet; the epic that adds one replaces this.
func (s *Service) recordLink(ctx context.Context, kind string, personID uuid.UUID, token string, expires time.Time) {
	link := s.baseURL + "/set-password?token=" + url.QueryEscape(token)

	s.logger.LogAttrs(ctx, slog.LevelWarn,
		"a sign-in link was issued and could not be sent, because sending mail is not set up yet",
		slog.String("kind", kind),
		slog.String("person", personID.String()),
		slog.String("link", link),
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
