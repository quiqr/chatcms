// Package onboarding brings a customer's existing site into ChatCMS: it gives
// the project a key of its own, imports the site, says what arrived, and
// records how the site should be published.
package onboarding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/secrets"
	"github.com/quiqr/chatcms/internal/store"
)

// ErrNoKeyYet says a project has no deploy key, which is its own answer
// rather than a refusal from a remote.
var ErrNoKeyYet = errors.New("this project has no key yet, so it cannot reach its remote")

// Keys makes, stores and lends out a project's deploy key.
//
// Git reaches a remote by running ssh with a key read from a file, so a key
// has to be on disk while it is used. One directory is made for the purpose
// when the server starts, readable only by the user the server runs as, and
// removed when it stops. See the change that introduced this package for why
// this is the least bad option.
type Keys struct {
	db      *store.DB
	keeper  *secrets.Keeper
	dir     string
	mu      sync.Mutex
	written map[uuid.UUID]string
}

// NewKeys prepares the place decrypted keys are written to.
func NewKeys(db *store.DB, keeper *secrets.Keeper, stateDir string) (*Keys, error) {
	dir := filepath.Join(stateDir, "deploy-keys")

	// Remove whatever an earlier run left, then make it afresh with
	// owner-only permission.
	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("clearing the key directory failed: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("preparing the key directory failed: %w", err)
	}

	return &Keys{db: db, keeper: keeper, dir: dir, written: map[uuid.UUID]string{}}, nil
}

// Dir is where decrypted keys are written, which the tests inspect.
func (k *Keys) Dir() string { return k.dir }

// Close removes every decrypted key.
func (k *Keys) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.written = map[uuid.UUID]string{}
	return os.RemoveAll(k.dir)
}

// Generate makes a new key for a project, replacing any it had, and returns
// the half an implementer gives to the git host.
func (k *Keys) Generate(ctx context.Context, projectID uuid.UUID, comment string) (string, error) {
	pair, err := secrets.GenerateKeyPair(comment)
	if err != nil {
		return "", err
	}

	sealed, err := k.keeper.SealString(pair.Private)
	if err != nil {
		return "", err
	}

	if err := k.db.Queries.SetProjectDeployKey(ctx, store.SetProjectDeployKeyParams{
		ID:               pgtype.UUID{Bytes: projectID, Valid: true},
		DeployPublicKey:  pair.Public,
		DeployPrivateKey: sealed,
	}); err != nil {
		return "", fmt.Errorf("storing the key failed: %w", err)
	}

	// A replaced key must not keep answering from the file an earlier one was
	// written to.
	k.forget(projectID)

	return pair.Public, nil
}

// PublicKey is the half that goes to the git host.
func (k *Keys) PublicKey(ctx context.Context, projectID uuid.UUID) (string, error) {
	row, err := k.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return "", fmt.Errorf("reading the project failed: %w", err)
	}
	if row.DeployPublicKey == "" {
		return "", ErrNoKeyYet
	}
	return row.DeployPublicKey, nil
}

// PathFor writes a project's key where git can read it, and returns that
// path. The file is written once per project per run.
func (k *Keys) PathFor(ctx context.Context, projectID uuid.UUID) (string, error) {
	k.mu.Lock()
	if path, ok := k.written[projectID]; ok {
		k.mu.Unlock()
		return path, nil
	}
	k.mu.Unlock()

	row, err := k.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return "", fmt.Errorf("reading the project failed: %w", err)
	}
	if len(row.DeployPrivateKey) == 0 {
		return "", ErrNoKeyYet
	}

	private, err := k.keeper.OpenString(row.DeployPrivateKey)
	if err != nil {
		return "", err
	}

	path := filepath.Join(k.dir, projectID.String())
	if err := os.WriteFile(path, []byte(private), 0o600); err != nil {
		return "", fmt.Errorf("writing the key failed: %w", err)
	}

	k.mu.Lock()
	k.written[projectID] = path
	k.mu.Unlock()

	return path, nil
}

func (k *Keys) forget(projectID uuid.UUID) {
	k.mu.Lock()
	path, ok := k.written[projectID]
	delete(k.written, projectID)
	k.mu.Unlock()

	if ok {
		os.Remove(path)
	}
}

// PathForSite is the lookup internal/gitsync is given: it takes a site
// identifier as text, because that layer knows nothing about projects.
//
// A project with no key gets no key path rather than an error: a remote on
// the same machine needs none, and a remote that does need one refuses, which
// the connection test turns into the message about adding the key.
func (k *Keys) PathForSite(ctx context.Context, siteID string) (string, error) {
	projectID, err := uuid.Parse(siteID)
	if err != nil {
		return "", fmt.Errorf("%q is not a project: %w", siteID, err)
	}

	path, err := k.PathFor(ctx, projectID)
	if errors.Is(err, ErrNoKeyYet) {
		return "", nil
	}
	return path, err
}

// Has says whether a project has a key yet.
func (k *Keys) Has(ctx context.Context, projectID uuid.UUID) (bool, error) {
	row, err := k.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return false, fmt.Errorf("reading the project failed: %w", err)
	}
	return len(row.DeployPrivateKey) > 0, nil
}
