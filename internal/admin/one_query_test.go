package admin

import (
	"os"
	"strings"
	"testing"
)

// TestTheListOfPeopleCostsOneQuery holds the rule that where somebody
// belongs is read with the list. A list of fifty people that asks the
// database once per person is a hundred and fifty requests on a screen that
// looks like one, and it only shows up once a customer has many people.
func TestTheListOfPeopleCostsOneQuery(t *testing.T) {
	source, err := os.ReadFile("admin.go")
	if err != nil {
		t.Fatal(err)
	}

	body := bodyOf(t, string(source), "func (s *Store) People(")
	if asked := strings.Count(body, "s.db.Queries."); asked != 1 {
		t.Errorf("Store.People asks the database %d times, want once", asked)
	}
	loop := strings.Index(body, "\n\tfor ")
	if loop >= 0 && strings.Contains(body[loop:], "s.db.Queries.") {
		t.Error("Store.People asks the database inside a loop")
	}
}

// bodyOf is the text of one function, from its opening line to the closing
// brace in the first column.
func bodyOf(t *testing.T, source, opening string) string {
	t.Helper()

	start := strings.Index(source, opening)
	if start < 0 {
		t.Fatalf("%q is not in the source", opening)
	}
	rest := source[start:]
	end := strings.Index(rest, "\n}\n")
	if end < 0 {
		t.Fatalf("%q does not end", opening)
	}
	return rest[:end]
}
