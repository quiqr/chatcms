//go:build integration

package tenancy_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/wholesite"
)

// withASite gives the harness somewhere to keep working copies, and makes
// one with a page in it, so that copying has something to copy.
func (h *harness) withASite(t *testing.T, project tenancy.Project, page string) *gitsync.Sites {
	t.Helper()

	cfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	repository, err := sites.Create(t.Context(), project.ID.String(), "", fstest.MapFS{
		"hugo.toml":             {Data: []byte("title = \"Korenaar\"\n")},
		"content/_index.md":     {Data: []byte(page)},
		"quiqr/model/base.yaml": {Data: []byte("collections: []\n")},
	}, gitsync.Identity{Name: "Joris", Email: "joris@example.org"}, "The site")
	if err != nil {
		t.Fatalf("making the site failed: %v", err)
	}
	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}

	h.service.MakesCopies(wholesite.NewCopier(sites))
	return sites
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

func TestACopySharesNothingWithItsOriginal(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	sites := h.withASite(t, project, "# Korenaar\n")

	made, err := h.service.Duplicate(t.Context(), h.who(admin), "technative", "korenaar",
		"korenaar-copy", "Korenaar copy", false, false)
	if err != nil {
		t.Fatalf("Duplicate() error = %v, want nil", err)
	}
	if made.Slug != "korenaar-copy" {
		t.Fatalf("the copy is %+v", made)
	}

	copied, err := h.store.ProjectBySlug(t.Context(), "technative", "korenaar-copy")
	if err != nil {
		t.Fatal(err)
	}

	// The copy holds what the original held...
	theCopy, err := sites.Open(t.Context(), copied.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = theCopy.Close() }()

	page, err := fs.ReadFile(theCopy.FS(), "content/_index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(page) != "# Korenaar\n" {
		t.Errorf("the copy holds %q", page)
	}

	// ...and changing the copy leaves the original as it was.
	if _, err := theCopy.Save(t.Context(), gitsync.Save{
		Writes:  map[string][]byte{"content/_index.md": []byte("# Only in the copy\n")},
		Author:  gitsync.Identity{Name: "Joris", Email: "joris@example.org"},
		Message: "Changed the copy",
	}); err != nil {
		t.Fatal(err)
	}

	original, err := sites.Open(t.Context(), project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = original.Close() }()

	still, err := fs.ReadFile(original.FS(), "content/_index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(still) != "# Korenaar\n" {
		t.Errorf("changing the copy changed the original: %q", still)
	}
}

func TestACopyPublishesNowhereUnlessItWasAskedFor(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	sites := h.withASite(t, project, "# Korenaar\n")

	if _, err := h.service.Duplicate(t.Context(), h.who(admin), "technative", "korenaar",
		"korenaar-copy", "Korenaar copy", false, false); err != nil {
		t.Fatal(err)
	}

	copied, err := h.store.ProjectBySlug(t.Context(), "technative", "korenaar-copy")
	if err != nil {
		t.Fatal(err)
	}

	// A copy with a remote would overwrite a live website the first time
	// anybody pressed Publish.
	where, err := sites.Dir(copied.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	settings, err := os.ReadFile(filepath.Join(where, ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(settings), "[remote ") {
		t.Errorf("the copy publishes somewhere:\n%s", settings)
	}

	// And it carries none of the original's history.
	theCopy, err := sites.Open(t.Context(), copied.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = theCopy.Close() }()

	versions, err := theCopy.History(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Errorf("the copy has %d versions, want the one that made it", len(versions))
	}
}

func TestACopyNeedsItsOwnName(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	_, err := h.service.Duplicate(t.Context(), h.who(admin), "technative", "korenaar",
		"korenaar-copy", project.Name, false, false)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("copying under the original's name error = %v, want a refusal", err)
	}
}

func TestACopyIsRefusedWhenThereIsNoRoomForASite(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	one := 1
	if err := h.store.SetLimits(t.Context(), organisation.ID, nil, &one); err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.Duplicate(t.Context(), h.who(admin), "technative", "korenaar",
		"korenaar-copy", "Korenaar copy", false, false); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("copying at the limit error = %v, want a refusal", err)
	}

	// And nothing was made on the way to refusing.
	if _, err := h.store.ProjectBySlug(t.Context(), "technative", "korenaar-copy"); err == nil {
		t.Error("the copy was made anyway")
	}
}

func TestAMovedSiteKeepsItsPeopleAndItsFiles(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	here := h.organisation(t, "technative", admin)
	there := h.organisation(t, "bakkerij", admin)
	project := h.project(t, here, "korenaar")
	sites := h.withASite(t, project, "# Korenaar\n")

	editor := h.person(t, "editor@example.org")
	if err := h.store.SetProjectRole(t.Context(), project.ID, editor.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}

	if err := h.service.MoveTo(t.Context(), h.who(admin), "technative", "korenaar", "bakkerij"); err != nil {
		t.Fatalf("MoveTo() error = %v, want nil", err)
	}

	// It is in the other organisation now, under the same address within it.
	if _, err := h.store.ProjectBySlug(t.Context(), "bakkerij", "korenaar"); err != nil {
		t.Errorf("the site is not in the organisation it moved to: %v", err)
	}
	if _, err := h.store.ProjectBySlug(t.Context(), "technative", "korenaar"); err == nil {
		t.Error("the site is still in the organisation it moved out of")
	}

	// The people kept their roles, and became members of the destination.
	members, err := h.store.Members(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	kept := false
	for _, member := range members {
		if member.PersonID == editor.ID && member.Role == policy.RoleEditor {
			kept = true
		}
	}
	if !kept {
		t.Errorf("the editor is now %+v", members)
	}
	already, err := h.store.IsInOrganisation(t.Context(), there.ID, editor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !already {
		t.Error("the editor did not become a member of the organisation the site moved to")
	}

	// Nothing on disk moved: a working copy is named by the project.
	theSite, err := sites.Open(t.Context(), project.ID.String())
	if err != nil {
		t.Fatalf("the files moved with it: %v", err)
	}
	_ = theSite.Close()
}

func TestMovingNeedsBothOrganisations(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	here := h.organisation(t, "technative", admin)
	there := h.organisation(t, "bakkerij", h.person(t, "somebody@example.org"))
	project := h.project(t, here, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	// A member of the destination is not an admin of it.
	if err := h.store.SetOrganisationRole(t.Context(), there.ID, admin.ID,
		policy.OrganisationMember); err != nil {
		t.Fatal(err)
	}

	if err := h.service.MoveTo(t.Context(), h.who(admin), "technative", "korenaar",
		"bakkerij"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("moving into an organisation they only belong to error = %v, want ErrNotAllowed", err)
	}
}

func TestMovingIsRefusedWhenThePeopleWouldNotFit(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	here := h.organisation(t, "technative", admin)
	there := h.organisation(t, "bakkerij", admin)
	project := h.project(t, here, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	for _, email := range []string{"one@example.org", "two@example.org"} {
		person := h.person(t, email)
		if err := h.store.SetProjectRole(t.Context(), project.ID, person.ID, policy.RoleEditor); err != nil {
			t.Fatal(err)
		}
	}

	// The destination holds its admin and has room for nobody else.
	one := 1
	if err := h.store.SetLimits(t.Context(), there.ID, &one, nil); err != nil {
		t.Fatal(err)
	}

	err := h.service.MoveTo(t.Context(), h.who(admin), "technative", "korenaar", "bakkerij")
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("moving with people that do not fit error = %v, want a refusal", err)
	}
	if refused.Key != "refused.no-room-for-the-people" {
		t.Errorf("the refusal is named %q", refused.Key)
	}

	// And nothing moved.
	if _, err := h.store.ProjectBySlug(t.Context(), "technative", "korenaar"); err != nil {
		t.Errorf("the site moved anyway: %v", err)
	}
}

func TestArchivingStopsASiteCountingAndBringsItBack(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	one := 1
	if err := h.store.SetLimits(t.Context(), organisation.ID, nil, &one); err != nil {
		t.Fatal(err)
	}

	if err := h.service.Archive(t.Context(), h.who(admin), "technative", "korenaar", true); err != nil {
		t.Fatalf("Archive() error = %v, want nil", err)
	}

	usage, err := h.store.UsageFor(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Projects.Held != 0 {
		t.Errorf("an archived site still counts: %+v", usage.Projects)
	}

	// Which is what makes room for another.
	if _, err := h.service.CreateProject(t.Context(), h.who(admin), "technative",
		"another", "Another", ""); err != nil {
		t.Errorf("making a site after archiving one error = %v, want nil", err)
	}

	// Bringing it back is refused while that room is taken, rather than
	// taking the organisation over its limit quietly.
	err = h.service.Archive(t.Context(), h.who(admin), "technative", "korenaar", false)
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("bringing it back over the limit error = %v, want a refusal", err)
	}
}

func TestTheZipHoldsTheSiteAndNothingOfGit(t *testing.T) {
	h := newHarness(t)
	admin := h.person(t, "admin@example.org")
	organisation := h.organisation(t, "technative", admin)
	project := h.project(t, organisation, "korenaar")
	h.withASite(t, project, "# Korenaar\n")

	var into bytes.Buffer
	name, err := h.service.WriteZip(t.Context(), h.who(admin), "technative", "korenaar", &into)
	if err != nil {
		t.Fatalf("WriteZip() error = %v, want nil", err)
	}
	if name != "korenaar" {
		t.Errorf("the zip is called %q", name)
	}

	opened, err := zip.NewReader(bytes.NewReader(into.Bytes()), int64(into.Len()))
	if err != nil {
		t.Fatal(err)
	}
	held := map[string]bool{}
	for _, file := range opened.File {
		held[file.Name] = true
		if len(file.Name) >= 4 && file.Name[:4] == ".git" {
			t.Errorf("the zip holds %q, which is how git keeps its own state", file.Name)
		}
	}
	for _, wanted := range []string{"content/_index.md", "hugo.toml", "quiqr/model/base.yaml"} {
		if !held[wanted] {
			t.Errorf("the zip does not hold %q: %v", wanted, held)
		}
	}
}
