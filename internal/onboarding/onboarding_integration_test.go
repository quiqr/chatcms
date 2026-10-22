//go:build integration

package onboarding_test

import (
	"crypto/rand"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/onboarding"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/secrets"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

const aModel = `singles:
  - key: home
    title: Homepage
    file: content/_index.md
    fields:
      - key: title
        type: string
`

type harness struct {
	db       *store.DB
	keys     *onboarding.Keys
	sites    *gitsync.Sites
	importer *onboarding.Importer
	gitCfg   gitsync.Config
	store    *tenancy.Store
	owner    auth.Person
	stateDir string
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

	masterKey := make([]byte, secrets.KeyLength)
	if _, err := rand.Read(masterKey); err != nil {
		t.Fatal(err)
	}
	keeper, err := secrets.NewKeeperFromBytes(masterKey)
	if err != nil {
		t.Fatal(err)
	}

	stateDir := t.TempDir()
	keys, err := onboarding.NewKeys(db, keeper, stateDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { keys.Close() })

	gitCfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(gitCfg, filepath.Join(stateDir, "sites"), keys.PathForSite)

	identity := auth.NewService(db, quiet(), "https://chatcms.example")
	owner, err := identity.People().Create(t.Context(), "owner@example.org", "Owner", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}

	return &harness{
		db:       db,
		keys:     keys,
		sites:    sites,
		importer: onboarding.NewImporter(db, sites, keys, gitCfg.System),
		gitCfg:   gitCfg,
		store:    tenancy.NewStore(db),
		owner:    owner,
		stateDir: stateDir,
	}
}

func (h *harness) project(t *testing.T, slug string) tenancy.Project {
	t.Helper()

	organisation, err := h.store.OrganisationBySlug(t.Context(), "technative")
	if errors.Is(err, tenancy.ErrNotFound) {
		organisation, err = h.store.CreateOrganisation(t.Context(), "technative", "TechNative", h.owner.ID)
	}
	if err != nil {
		t.Fatal(err)
	}

	project, err := h.store.CreateProject(t.Context(), organisation.ID, slug, slug, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.store.SetProjectRole(t.Context(), project.ID, h.owner.ID, policy.RoleOwner); err != nil {
		t.Fatal(err)
	}
	return project
}

// aSite is a Hugo site with a Quiqr model, as a customer's would be.
func aSite(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	write(t, filepath.Join(dir, "hugo.toml"), "baseURL = 'https://korenaar.example'\n")
	write(t, filepath.Join(dir, "quiqr", "model", "base.yaml"), aModel)
	write(t, filepath.Join(dir, "content", "_index.md"), "---\ntitle: Home\n---\n\nWelcome.\n")
	return dir
}

func bareRemote(t *testing.T, cfg gitsync.Config) string {
	t.Helper()

	remote := filepath.Join(t.TempDir(), "remote.git")
	gitRun(t, cfg, "", "init", "--bare", "--initial-branch=main", remote)
	return remote
}

// remoteHolding makes a remote whose content is the given site.
func remoteHolding(t *testing.T, cfg gitsync.Config, site string) string {
	t.Helper()

	remote := bareRemote(t, cfg)

	seed := filepath.Join(t.TempDir(), "seed")
	gitRun(t, cfg, "", "clone", "--", remote, seed)
	copyTree(t, site, seed)
	gitRun(t, cfg, seed, "add", "--all")
	gitRun(t, cfg, seed, "commit", "--message=The site")
	gitRun(t, cfg, seed, "push", "origin", "HEAD:refs/heads/main")

	return remote
}

func TestImportFromARepository(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")
	remote := remoteHolding(t, h.gitCfg, aSite(t))

	result, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceRepository, Address: remote,
	})
	if err != nil {
		t.Fatalf("Import() error = %v, want nil", err)
	}

	if !result.Inspection.Ready() {
		t.Fatalf("inspection = %+v, want a Hugo site with a model", result.Inspection)
	}
	if result.Inspection.Pages != 1 {
		t.Errorf("pages = %d, want 1", result.Inspection.Pages)
	}
	if result.Remote != remote {
		t.Errorf("remote = %q, want the address it was imported from", result.Remote)
	}
}

func TestImportFromAnExport(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")
	remote := bareRemote(t, h.gitCfg)

	result, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceExport, Address: aSite(t), Remote: remote,
	})
	if err != nil {
		t.Fatalf("Import() error = %v, want nil", err)
	}
	if !result.Inspection.Ready() {
		t.Fatalf("inspection = %+v, want a Hugo site with a model", result.Inspection)
	}

	// The content is in the working copy and a version was made of it.
	repository, err := h.sites.Open(t.Context(), project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	if _, err := fs.ReadFile(repository.FS(), "content/_index.md"); err != nil {
		t.Errorf("the exported content is not in the working copy: %v", err)
	}

	versions, err := repository.History(t.Context(), 10)
	if err != nil || len(versions) == 0 {
		t.Fatalf("history = %+v, %v, want the import recorded as a version", versions, err)
	}
}

func TestAnExportAndItsRepositoryAgree(t *testing.T) {
	h := newHarness(t)
	site := aSite(t)

	fromExport := h.project(t, "from-export")
	if _, err := h.importer.Import(t.Context(), fromExport.ID, onboarding.Request{
		Kind: onboarding.SourceExport, Address: site, Remote: bareRemote(t, h.gitCfg),
	}); err != nil {
		t.Fatal(err)
	}

	fromRepository := h.project(t, "from-repository")
	if _, err := h.importer.Import(t.Context(), fromRepository.ID, onboarding.Request{
		Kind: onboarding.SourceRepository, Address: remoteHolding(t, h.gitCfg, site),
	}); err != nil {
		t.Fatal(err)
	}

	first := contentOf(t, h, fromExport.ID)
	second := contentOf(t, h, fromRepository.ID)

	if len(first) == 0 {
		t.Fatal("the imported site has no content")
	}
	if len(first) != len(second) {
		t.Fatalf("the two imports hold %d and %d files", len(first), len(second))
	}
	for name, contents := range first {
		if second[name] != contents {
			t.Errorf("%s differs between the export and the repository", name)
		}
	}
}

func TestImportFromATemplate(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	onboarding.Templates["a-bakery"] = remoteHolding(t, h.gitCfg, aSite(t))
	t.Cleanup(func() { delete(onboarding.Templates, "a-bakery") })

	result, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "a-bakery",
	})
	if err != nil {
		t.Fatalf("Import() error = %v, want nil", err)
	}
	if !result.Inspection.Ready() {
		t.Errorf("inspection = %+v", result.Inspection)
	}

	if _, err := h.importer.Import(t.Context(), h.project(t, "other").ID, onboarding.Request{
		Kind: onboarding.SourceTemplate, Address: "no-such-template",
	}); !errors.Is(err, onboarding.ErrNoSuchTemplate) {
		t.Errorf("an unknown template gave %v, want ErrNoSuchTemplate", err)
	}
}

func TestProjectRecordsItsSource(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	before, err := h.db.Queries.GetProject(t.Context(), pgUUID(project.ID))
	if err != nil {
		t.Fatal(err)
	}
	if before.SourceKind != string(onboarding.SourceNone) {
		t.Errorf("a project with nothing imported says %q", before.SourceKind)
	}

	remote := remoteHolding(t, h.gitCfg, aSite(t))
	if _, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
		Kind: onboarding.SourceRepository, Address: remote,
	}); err != nil {
		t.Fatal(err)
	}

	after, err := h.db.Queries.GetProject(t.Context(), pgUUID(project.ID))
	if err != nil {
		t.Fatal(err)
	}
	if after.SourceKind != string(onboarding.SourceRepository) || after.SourceAddress != remote {
		t.Errorf("source = %q %q", after.SourceKind, after.SourceAddress)
	}
}

func TestKeysDifferPerProjectAndArePrivate(t *testing.T) {
	h := newHarness(t)
	first := h.project(t, "first")
	second := h.project(t, "second")

	firstKey, err := h.keys.Generate(t.Context(), first.ID, "chatcms first")
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := h.keys.Generate(t.Context(), second.ID, "chatcms second")
	if err != nil {
		t.Fatal(err)
	}

	if firstKey == secondKey {
		t.Error("two projects share a key")
	}
	for _, key := range []string{firstKey, secondKey} {
		if !strings.HasPrefix(key, "ssh-ed25519 ") {
			t.Errorf("public key = %q", key)
		}
	}

	// What is stored is sealed.
	row, err := h.db.Queries.GetProject(t.Context(), pgUUID(first.ID))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(row.DeployPrivateKey), "OPENSSH") {
		t.Error("the stored private key is readable")
	}
}

func TestKeyFileIsOwnerOnlyAndGoesAwayOnClose(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}

	directory, err := os.Stat(h.keys.Dir())
	if err != nil {
		t.Fatal(err)
	}
	if directory.Mode().Perm() != 0o700 {
		t.Errorf("the key directory is %v, want owner-only", directory.Mode().Perm())
	}

	path, err := h.keys.PathFor(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if file.Mode().Perm() != 0o600 {
		t.Errorf("the key file is %v, want owner-only", file.Mode().Perm())
	}

	if err := h.keys.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(h.keys.Dir()); !errors.Is(err, os.ErrNotExist) {
		t.Error("the key directory is still there after closing")
	}
}

func TestReplacingAKeyForgetsTheOldFile(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}
	path, err := h.keys.PathFor(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	replaced, err := h.keys.Generate(t.Context(), project.ID, "chatcms")
	if err != nil {
		t.Fatal(err)
	}

	again, err := h.keys.PathFor(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(again)
	if err != nil {
		t.Fatal(err)
	}

	if string(first) == string(second) {
		t.Error("the replaced key still answers from the old file")
	}
	if !strings.HasPrefix(replaced, "ssh-ed25519 ") {
		t.Errorf("the replacement public key = %q", replaced)
	}
}

func TestNoKeyYetIsItsOwnMessage(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if _, err := h.keys.PathFor(t.Context(), project.ID); !errors.Is(err, onboarding.ErrNoKeyYet) {
		t.Errorf("PathFor() with no key = %v, want ErrNoKeyYet", err)
	}
	if _, err := h.keys.PublicKey(t.Context(), project.ID); !errors.Is(err, onboarding.ErrNoKeyYet) {
		t.Errorf("PublicKey() with no key = %v, want ErrNoKeyYet", err)
	}

	connection, err := h.importer.TestConnection(t.Context(), project.ID, "git@example.org:a/b.git")
	if err != nil {
		t.Fatal(err)
	}
	if connection.Good || !strings.Contains(connection.Message, "no key yet") {
		t.Errorf("connection = %+v, want it to say the project has no key", connection)
	}
}

func TestConnectionIsGoodAgainstALocalRemote(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")
	remote := remoteHolding(t, h.gitCfg, aSite(t))

	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}

	connection, err := h.importer.TestConnection(t.Context(), project.ID, remote)
	if err != nil {
		t.Fatal(err)
	}
	if !connection.Good {
		t.Errorf("connection = %+v, want it to be good", connection)
	}
	if connection.PublicKey != "" {
		t.Error("a good connection showed the key, which only a refusal should")
	}
}

func TestRemoteUnreachable(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}

	connection, err := h.importer.TestConnection(t.Context(), project.ID,
		filepath.Join(t.TempDir(), "no-such-remote.git"))
	if err != nil {
		t.Fatal(err)
	}
	if connection.Good {
		t.Error("a remote that is not there was reported as good")
	}
	if connection.Message == "" {
		t.Error("nothing was said about what to do")
	}
}

func TestConnectionAnswersCarryNoSecret(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if _, err := h.keys.Generate(t.Context(), project.ID, "chatcms"); err != nil {
		t.Fatal(err)
	}

	for _, remote := range []string{
		remoteHolding(t, h.gitCfg, aSite(t)),
		filepath.Join(t.TempDir(), "absent.git"),
		"",
	} {
		connection, err := h.importer.TestConnection(t.Context(), project.ID, remote)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(connection.Message, "PRIVATE KEY") {
			t.Errorf("the answer carries a private key: %q", connection.Message)
		}
		if connection.PublicKey != "" && strings.Contains(connection.PublicKey, "PRIVATE") {
			t.Error("the key shown is not the public half")
		}
	}
}

func TestPublishingSettings(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
		To: onboarding.PublishToItsOwnLine, Form: onboarding.PublishBuilt,
	}); err != nil {
		t.Fatalf("SetPublishing() error = %v, want nil", err)
	}

	row, err := h.db.Queries.GetProject(t.Context(), pgUUID(project.ID))
	if err != nil {
		t.Fatal(err)
	}
	if row.PublishBranch != onboarding.DefaultPublishBranch {
		t.Errorf("branch = %q, want the default offered", row.PublishBranch)
	}
	if row.PublishForm != string(onboarding.PublishBuilt) {
		t.Errorf("form = %q", row.PublishForm)
	}

	if err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
		To: onboarding.PublishToMainLine, Branch: "ignored", Form: onboarding.PublishSource,
	}); err != nil {
		t.Fatal(err)
	}
	row, _ = h.db.Queries.GetProject(t.Context(), pgUUID(project.ID))
	if row.PublishBranch != "" {
		t.Errorf("branch = %q, want none when publishing to the main line", row.PublishBranch)
	}
}

func TestRefusesAnUnusableBranchName(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	for _, name := range []string{"-leading", "trailing-", "has space", "double//slash", "has~tilde"} {
		err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
			To: onboarding.PublishToItsOwnLine, Branch: name, Form: onboarding.PublishSource,
		})
		if !errors.Is(err, onboarding.ErrUnusableBranchName) {
			t.Errorf("SetPublishing(%q) error = %v, want ErrUnusableBranchName", name, err)
		}
	}

	for _, name := range []string{"quiqr/publish", "main", "release-2026", "a_b"} {
		if err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
			To: onboarding.PublishToItsOwnLine, Branch: name, Form: onboarding.PublishSource,
		}); err != nil {
			t.Errorf("SetPublishing(%q) error = %v, want nil", name, err)
		}
	}
}

func TestRefusesAnUnknownPublishingChoice(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	if err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
		To: "somewhere-else", Form: onboarding.PublishSource,
	}); err == nil {
		t.Error("an unknown place to publish was accepted")
	}
	if err := h.importer.SetPublishing(t.Context(), project.ID, onboarding.Publishing{
		To: onboarding.PublishToMainLine, Form: "something-else",
	}); err == nil {
		t.Error("an unknown form was accepted")
	}
}

func TestImportFromAnUnreadableExport(t *testing.T) {
	h := newHarness(t)
	project := h.project(t, "korenaar")

	for _, path := range []string{"", filepath.Join(t.TempDir(), "absent")} {
		if _, err := h.importer.Import(t.Context(), project.ID, onboarding.Request{
			Kind: onboarding.SourceExport, Address: path,
		}); !errors.Is(err, onboarding.ErrExportUnreadable) {
			t.Errorf("Import(%q) error = %v, want ErrExportUnreadable", path, err)
		}
	}
}

// serviceFor builds the layer the HTTP endpoints talk to, with its own
// project and people, so that what each role may do can be checked.
func serviceFor(t *testing.T, h *harness) (*onboarding.Service, tenancy.Project, map[policy.Role]httpapi.Actor) {
	t.Helper()

	service := onboarding.NewService(h.db, h.store, h.keys, h.importer)
	project := h.project(t, "korenaar")

	identity := auth.NewService(h.db, quiet(), "https://chatcms.example")
	actors := map[policy.Role]httpapi.Actor{}

	for _, role := range policy.Roles {
		person, err := identity.People().Create(t.Context(),
			"the-"+string(role)+"@example.org", string(role), "a long enough passphrase")
		if err != nil {
			t.Fatal(err)
		}
		if err := h.store.SetProjectRole(t.Context(), project.ID, person.ID, role); err != nil {
			t.Fatal(err)
		}
		actors[role] = httpapi.Actor{ID: person.ID.String(), Email: person.Email}
	}

	return service, project, actors
}

func TestOnlyThoseWhoMayChangeSettingsMay(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)

	mayChange := map[policy.Role]bool{
		policy.RoleOwner: true, policy.RoleImplementer: true,
		policy.RoleEditor: false, policy.RoleViewer: false,
	}

	for role, allowed := range mayChange {
		t.Run(string(role), func(t *testing.T) {
			who := actors[role]

			_, keyErr := service.GenerateKey(t.Context(), who, "technative", "korenaar")
			_, publishErr := service.SetPublishing(t.Context(), who, "technative", "korenaar",
				"main-line", "", "source")
			_, importErr := service.ImportSite(t.Context(), who, "technative", "korenaar",
				"template", "no-such-template", "")

			if allowed {
				if errors.Is(keyErr, httpapi.ErrNotAllowed) {
					t.Errorf("%s may not make a key", role)
				}
				if errors.Is(publishErr, httpapi.ErrNotAllowed) {
					t.Errorf("%s may not set publishing", role)
				}
				if errors.Is(importErr, httpapi.ErrNotAllowed) {
					t.Errorf("%s may not import", role)
				}
				return
			}

			for what, err := range map[string]error{
				"make a key":     keyErr,
				"set publishing": publishErr,
				"import":         importErr,
			} {
				if !errors.Is(err, httpapi.ErrNotAllowed) {
					t.Errorf("%s may %s: %v", role, what, err)
				}
			}
		})
	}
}

func TestSeeingTheKeyIsForImplementersOnly(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)

	if _, err := service.GenerateKey(t.Context(), actors[policy.RoleOwner], "technative", "korenaar"); err != nil {
		t.Fatal(err)
	}

	owner, err := service.ReadKey(t.Context(), actors[policy.RoleOwner], "technative", "korenaar")
	if err != nil {
		t.Fatalf("an owner cannot read the key: %v", err)
	}
	if !strings.HasPrefix(owner.PublicKey, "ssh-ed25519 ") {
		t.Errorf("key = %q", owner.PublicKey)
	}

	if _, err := service.ReadKey(t.Context(), actors[policy.RoleEditor], "technative", "korenaar"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("an editor reading the key gave %v, want ErrNotAllowed", err)
	}
	if _, err := service.TestConnection(t.Context(), actors[policy.RoleViewer], "technative", "korenaar"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("a viewer testing the connection gave %v, want ErrNotAllowed", err)
	}
}

func TestReadingAProjectThatIsNotYours(t *testing.T) {
	h := newHarness(t)
	service, _, _ := serviceFor(t, h)

	identity := auth.NewService(h.db, quiet(), "https://chatcms.example")
	stranger, err := identity.People().Create(t.Context(), "stranger@example.org", "Stranger", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	who := httpapi.Actor{ID: stranger.ID.String(), Email: stranger.Email}

	for name, call := range map[string]func() error{
		"read the key":    func() error { _, err := service.ReadKey(t.Context(), who, "technative", "korenaar"); return err },
		"inspect":         func() error { _, err := service.InspectSite(t.Context(), who, "technative", "korenaar"); return err },
		"read publishing": func() error { _, err := service.ReadPublishing(t.Context(), who, "technative", "korenaar"); return err },
	} {
		if err := call(); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("a stranger could %s: %v", name, err)
		}
	}

	// A project that does not exist answers the same way.
	if _, err := service.ReadKey(t.Context(), who, "technative", "no-such-project"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("a project that is not there gave %v, want ErrMissing", err)
	}
}

func TestNobodyReachesOnboarding(t *testing.T) {
	h := newHarness(t)
	service, _, _ := serviceFor(t, h)

	if _, err := service.ReadKey(t.Context(), httpapi.Anonymous(), "technative", "korenaar"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("nobody read a key: %v", err)
	}
}

func TestTheServiceReportsWhatArrived(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)
	who := actors[policy.RoleImplementer]

	before, err := service.InspectSite(t.Context(), who, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if before.Kind != "none" {
		t.Errorf("a project with nothing imported says %q", before.Kind)
	}

	remote := remoteHolding(t, h.gitCfg, aSite(t))
	imported, err := service.ImportSite(t.Context(), who, "technative", "korenaar", "repository", remote, "")
	if err != nil {
		t.Fatalf("ImportSite() error = %v, want nil", err)
	}
	if !imported.Ready || imported.Pages != 1 {
		t.Errorf("imported = %+v", imported)
	}

	after, err := service.InspectSite(t.Context(), who, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if !after.Ready || after.Remote != remote {
		t.Errorf("after = %+v", after)
	}

	connection, err := service.TestConnection(t.Context(), who, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if connection.Message == "" {
		t.Error("the connection answer says nothing")
	}
}

func TestTheServiceRefusesWhatCannotBeDone(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)
	who := actors[policy.RoleImplementer]

	if _, err := service.SetPublishing(t.Context(), who, "technative", "korenaar",
		"its-own-line", "-bad-name-", "source"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("an unusable name gave %v, want a refusal", err)
	}
	if _, err := service.ImportSite(t.Context(), who, "technative", "korenaar",
		"template", "no-such-template", ""); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("an unknown template gave %v, want a refusal", err)
	}
	if _, err := service.ImportSite(t.Context(), who, "technative", "korenaar",
		"export", "/no/such/place", ""); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("an unreadable export gave %v, want a refusal", err)
	}
}

func TestReadingAKeyBeforeThereIsOne(t *testing.T) {
	h := newHarness(t)
	service, _, actors := serviceFor(t, h)

	key, err := service.ReadKey(t.Context(), actors[policy.RoleOwner], "technative", "korenaar")
	if err != nil {
		t.Fatalf("ReadKey() error = %v, want nil", err)
	}
	if key.PublicKey != "" {
		t.Errorf("key = %q, want none", key.PublicKey)
	}
	if !strings.Contains(key.Advice, "no key yet") {
		t.Errorf("advice = %q, want it to say there is no key yet", key.Advice)
	}
}
