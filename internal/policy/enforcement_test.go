package policy

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// roleValue matches a role written as a value, which is what comparing one
// looks like. Prose about implementers, and a test helper called editor, are
// not comparisons and are left alone; only a quoted role is a role.
var roleValue = regexp.MustCompile(`"(owner|implementer|editor|viewer)"`)

// allowedToNameARole are the places a role may legitimately appear outside
// this package. They are listed rather than matched by pattern, so adding one
// is a visible change in a diff.
var allowedToNameARole = []string{
	// Stores roles and hands them to this package.
	"internal/tenancy/",
	// Generated from the SQL, which carries the column.
	"internal/store/",
	// The sign-in and shell tests describe people by role in their fixtures.
	"internal/app/",
	// Carries the role only as a value it has been told.
	"internal/httpapi/",
}

// TestOnlyPolicyNamesARole enforces the rule in docs/ARCHITECTURE.md: one
// function decides every permission, so no other package compares a role. A
// package that merely carries a role as a value is listed above.
func TestOnlyPolicyNamesARole(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Skipf("the repository root is not readable from here: %v", err)
	}

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".jj", "node_modules", "vendor", "result", ".direnv", ".pgdata", "web", "docs", "openspec":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)

		if strings.HasPrefix(name, "internal/policy/") {
			return nil
		}
		for _, exception := range allowedToNameARole {
			if strings.HasPrefix(name, exception) {
				return nil
			}
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if match := roleValue.Find(source); match != nil {
			t.Errorf("%s carries the role %s as a value; only internal/policy decides what a role may do",
				name, match)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the sources failed: %v", err)
	}
}

// administersField catches a package reading whether somebody administers the
// installation. Setting it where the actor is loaded is the one legitimate
// use; deciding anything from it belongs in Can.
var administersField = regexp.MustCompile(`\.Administers\b`)

// allowedToNameAdministering is where the flag may appear at all: the policy
// that decides with it, and the one place that loads an actor from the
// database.
var allowedToNameAdministering = []string{
	"internal/policy/",
	"internal/tenancy/tenancy.go",
	// Carries it only as a value it has been told, the same way it carries a
	// role: the answer a screen is given says whether the person looking
	// administers, so that the interface can leave out what it must. Nothing
	// here decides from it.
	"internal/httpapi/",
	// The list of people says, beside each one, whether they administer the
	// installation. These three read that column and carry it to the screen;
	// internal/admin/service.go, which is where a decision would be
	// tempting, is deliberately not listed and still asks policy.Can.
	"internal/admin/admin.go",
	"internal/admin/web.go",
	"internal/admin/admin_integration_test.go",
}

// generatedGo is code sqlc writes from the SQL. A column named in a struct it
// generated is the database's word for the field, not a decision made from
// it, and rewriting the generator is not the point of this rule.
var generatedGo = regexp.MustCompile(`\.generated\.go$`)

func TestOnlyPolicyDecidesFromAdministering(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Skipf("the repository root is not readable from here: %v", err)
	}

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".jj", "node_modules", "vendor", "result", ".direnv", ".pgdata", "web", "docs", "openspec":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if generatedGo.MatchString(name) {
			return nil
		}
		for _, exception := range allowedToNameAdministering {
			if strings.HasPrefix(name, exception) {
				return nil
			}
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if administersField.Match(source) {
			t.Errorf("%s decides from whether somebody administers the installation; ask policy.Can instead", name)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the sources failed: %v", err)
	}
}

func TestTheAdministeringRuleCatchesAnOffender(t *testing.T) {
	for _, offending := range []string{
		"if actor.Administers { allow() }",
		"return who.Administers",
	} {
		if !administersField.MatchString(offending) {
			t.Errorf("the check would not notice %q", offending)
		}
	}
	if administersField.MatchString("// an administrator may do this") {
		t.Error("the check trips on prose")
	}
}

func TestTheRoleRuleCatchesAnOffender(t *testing.T) {
	// The rule is worth nothing if it cannot fail, so prove the check does.
	for _, offending := range []string{
		`if member.Role == "editor" { allow() }`,
		`roles := []string{"owner", "implementer"}`,
		`case "viewer":`,
	} {
		if !roleValue.MatchString(offending) {
			t.Errorf("the check would not notice %q", offending)
		}
	}

	// Prose, identifiers and longer words are not comparisons.
	for _, innocent := range []string{
		"// a warning the implementer reads",
		"func editor(name string) Identity {",
		`"this is shown to the implementer only"`,
		"the implementers' guide to viewership",
		"OwnerName string",
	} {
		if roleValue.MatchString(innocent) {
			t.Errorf("the check trips on %q, which compares nothing", innocent)
		}
	}
}

func TestEveryExceptionIsStillNeeded(t *testing.T) {
	root, _ := filepath.Abs("../..")

	for _, exception := range allowedToNameARole {
		if _, err := os.Stat(filepath.Join(root, exception)); err != nil {
			t.Errorf("%s is listed as allowed to name a role but does not exist; remove it", exception)
		}
	}
}
