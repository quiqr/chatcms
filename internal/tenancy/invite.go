package tenancy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/policy"
)

var (
	// ErrNotARole says the role asked for is not one of the four.
	ErrNotARole = errors.New("that is not a role")
	// ErrNoAddress says an invitation arrived without an email address.
	ErrNoAddress = errors.New("an invitation needs an email address")
)

// Inviter brings someone onto a project. A person who does not exist yet is
// created with no password and given a link that sets one; the link is issued
// by internal/auth, which already owns single-use links.
type Inviter struct {
	store *Store
	auth  *auth.Service
}

func NewInviter(store *Store, service *auth.Service) *Inviter {
	return &Inviter{store: store, auth: service}
}

// Invitation says what happened, so the interface can tell the person who
// invited whether a link went out.
type Invitation struct {
	PersonID uuid.UUID
	Email    string
	Role     policy.Role
	// LinkIssued says a link that sets a first password was issued, which
	// happens for someone who has no password yet.
	LinkIssued bool
}

// Invite gives an address a role on a project.
func (i *Inviter) Invite(ctx context.Context, projectID uuid.UUID, email string, role policy.Role) (Invitation, error) {
	address := strings.ToLower(strings.TrimSpace(email))
	if address == "" {
		return Invitation{}, ErrNoAddress
	}
	if !knownRole(role) {
		return Invitation{}, fmt.Errorf("%w: %q", ErrNotARole, role)
	}

	person, err := i.auth.People().ByEmail(ctx, address)
	switch {
	case errors.Is(err, auth.ErrNoSuchPerson):
		person, err = i.auth.People().Create(ctx, address, "", "")
		if err != nil {
			return Invitation{}, err
		}
	case err != nil:
		return Invitation{}, err
	}

	if err := i.store.SetProjectRole(ctx, projectID, person.ID, role); err != nil {
		return Invitation{}, err
	}

	invitation := Invitation{PersonID: person.ID, Email: person.Email, Role: role}

	// Someone who has never set a password needs a way in; someone who has
	// one already simply gains the role.
	if !person.HasPassword {
		if err := i.auth.Invite(ctx, person.ID); err != nil {
			return Invitation{}, err
		}
		invitation.LinkIssued = true
	}

	return invitation, nil
}

func knownRole(role policy.Role) bool {
	for _, known := range policy.Roles {
		if known == role {
			return true
		}
	}
	return false
}
