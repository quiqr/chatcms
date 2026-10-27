package history_test

import (
	"strings"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/history"
)

func at(day int, hour int) time.Time {
	return time.Date(2026, 3, day, hour, 0, 0, 0, time.UTC)
}

var now = at(20, 12)

func aVersion(id string, when time.Time, message string, paths ...string) gitsync.Version {
	return gitsync.Version{
		ID:      id,
		Time:    when,
		Author:  gitsync.Identity{Name: "Anna"},
		Message: message,
		Paths:   paths,
	}
}

func TestTheNoteSomebodyWroteIsWhatTheVersionIsCalled(t *testing.T) {
	timeline := history.Build(
		[]gitsync.Version{aVersion("aaa", at(20, 9), "Save content/_index.md", "content/_index.md")},
		map[string]string{"aaa": "Spring prices"},
		"aaa", 0, now,
	)

	if got := timeline.Groups[0].Entries[0].Title; got != "Spring prices" {
		t.Errorf("Title = %q, want the note somebody wrote", got)
	}
}

func TestWithoutANoteAReadableMessageIsUsed(t *testing.T) {
	timeline := history.Build(
		[]gitsync.Version{aVersion("aaa", at(20, 9), "Fix the opening hours", "content/_index.md")},
		nil, "aaa", 0, now,
	)

	if got := timeline.Groups[0].Entries[0].Title; got != "Fix the opening hours" {
		t.Errorf("Title = %q, want the version's own message", got)
	}
}

func TestATechnicalMessageIsReplacedByASummary(t *testing.T) {
	timeline := history.Build(
		[]gitsync.Version{
			aVersion("aaa", at(20, 9), "Merge branch 'main' into main", "content/a.md", "content/b.md"),
		},
		nil, "aaa", 0, now,
	)

	title := timeline.Groups[0].Entries[0].Title
	if strings.Contains(strings.ToLower(title), "merge") || strings.Contains(title, "branch") {
		t.Errorf("Title = %q, which carries a word an editor should not see", title)
	}
	if title != "2 pages changed" {
		t.Errorf("Title = %q, want a summary of what changed", title)
	}
}

func TestASummarySaysWhatSortOfThingChanged(t *testing.T) {
	cases := map[string][]string{
		"1 page changed":                  {"content/_index.md"},
		"3 pages changed":                 {"content/a.md", "content/b.md", "content/c.md"},
		"1 image changed":                 {"static/logo.png"},
		"1 page and 2 images changed":     {"content/a.md", "static/a.png", "static/b.jpg"},
		"1 other file changed":            {"layouts/index.html"},
		"1 page and 1 other file changed": {"content/a.md", "layouts/index.html"},
	}

	for want, paths := range cases {
		timeline := history.Build(
			[]gitsync.Version{aVersion("aaa", at(20, 9), "Merge something", paths...)},
			nil, "aaa", 0, now,
		)
		if got := timeline.Groups[0].Entries[0].Title; got != want {
			t.Errorf("paths %v gave %q, want %q", paths, got, want)
		}
	}
}

func TestAVersionThatTouchedNothingStillReads(t *testing.T) {
	timeline := history.Build(
		[]gitsync.Version{aVersion("aaa", at(20, 9), "Merge")},
		nil, "aaa", 0, now,
	)

	if got := timeline.Groups[0].Entries[0].Title; got == "" {
		t.Error("a version with nothing in it has no title")
	}
}

func TestVersionsAreGroupedByPeriod(t *testing.T) {
	timeline := history.Build([]gitsync.Version{
		aVersion("today", at(20, 9), "Today's work", "content/a.md"),
		aVersion("yesterday", at(19, 9), "Yesterday's work", "content/a.md"),
		aVersion("thisweek", at(16, 9), "Earlier in the week", "content/a.md"),
		aVersion("thismonth", at(3, 9), "Earlier this month", "content/a.md"),
		aVersion("february", time.Date(2026, 2, 3, 9, 0, 0, 0, time.UTC), "In February", "content/a.md"),
		aVersion("lastyear", time.Date(2025, 11, 3, 9, 0, 0, 0, time.UTC), "Last year", "content/a.md"),
	}, nil, "today", 0, now)

	want := []string{"Today", "Yesterday", "Earlier this week", "Earlier this month", "February", "November 2025"}
	if len(timeline.Groups) != len(want) {
		t.Fatalf("groups = %v, want %v", periods(timeline), want)
	}
	for index, period := range want {
		if timeline.Groups[index].Period != period {
			t.Errorf("group %d = %q, want %q (all: %v)", index, timeline.Groups[index].Period, period, periods(timeline))
		}
	}
}

func periods(timeline history.Timeline) []string {
	out := []string{}
	for _, group := range timeline.Groups {
		out = append(out, group.Period)
	}
	return out
}

func TestTheMostRecentVersionComesFirst(t *testing.T) {
	timeline := history.Build([]gitsync.Version{
		aVersion("older", at(20, 9), "Older", "content/a.md"),
		aVersion("newer", at(20, 11), "Newer", "content/a.md"),
	}, nil, "newer", 0, now)

	entries := timeline.Groups[0].Entries
	if entries[0].Title != "Newer" {
		t.Errorf("the first entry is %q, want the most recent", entries[0].Title)
	}
}

func TestTheVersionTheLiveSiteHoldsIsMarked(t *testing.T) {
	timeline := history.Build([]gitsync.Version{
		aVersion("waiting", at(20, 11), "Not published yet", "content/a.md"),
		aVersion("live", at(20, 9), "Live now", "content/a.md"),
	}, nil, "live", 1, now)

	entries := timeline.Groups[0].Entries
	if entries[0].Live {
		t.Error("a version waiting to be published is marked as live")
	}
	if !entries[1].Live {
		t.Error("the version the live site holds is not marked")
	}
	if timeline.Unpublished != 1 {
		t.Errorf("Unpublished = %d, want 1", timeline.Unpublished)
	}
}

func TestATimelineCarriesNoVersionIdentifierForAPersonToRead(t *testing.T) {
	timeline := history.Build(
		[]gitsync.Version{aVersion("4b825dc642cb6eb9a060e54bf8d69288fbee4904", at(20, 9), "Some work", "content/a.md")},
		nil, "", 0, now,
	)

	entry := timeline.Groups[0].Entries[0]
	for _, shown := range []string{entry.Title, entry.By, entry.Note, timeline.Groups[0].Period} {
		if strings.Contains(shown, "4b825dc") {
			t.Errorf("%q carries a version identifier", shown)
		}
	}
	if entry.Version == "" {
		t.Error("the entry carries no identifier at all, so nothing could go back to it")
	}
}

func TestAnEmptyHistoryIsAnEmptyTimeline(t *testing.T) {
	timeline := history.Build(nil, nil, "", 0, now)

	if len(timeline.Groups) != 0 {
		t.Errorf("groups = %v, want none", periods(timeline))
	}
}

func TestAVersionWrittenInAnotherOffsetIsStillToday(t *testing.T) {
	// Git records the offset the commit was made in. A change made ten
	// minutes ago, just after midnight in that offset, must not be filed
	// under Yesterday because the server reads its clock in another.
	east := time.FixedZone("east", 2*60*60)
	justAfterMidnight := time.Date(2026, 10, 27, 0, 25, 0, 0, east)

	now := justAfterMidnight.In(time.UTC).Add(10 * time.Minute)

	timeline := history.Build(
		[]gitsync.Version{aVersionAt("just-now", justAfterMidnight, "A change")},
		nil, "", 0, now,
	)

	if said := timeline.Groups[0].Period; said != "Today" {
		t.Errorf("a change ten minutes old was filed under %q, want Today", said)
	}
}

func TestADayOldChangeIsYesterdayWhateverTheOffset(t *testing.T) {
	west := time.FixedZone("west", -7*60*60)
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	theDayBefore := time.Date(2026, 10, 26, 5, 0, 0, 0, west)

	timeline := history.Build(
		[]gitsync.Version{aVersionAt("then", theDayBefore, "A change")},
		nil, "", 0, now,
	)

	if said := timeline.Groups[0].Period; said != "Yesterday" {
		t.Errorf("a change from the day before was filed under %q, want Yesterday", said)
	}
}

// aVersionAt is one version at a given moment, which is all these two tests
// are about.
func aVersionAt(id string, when time.Time, title string) gitsync.Version {
	return gitsync.Version{
		ID:      id,
		Time:    when,
		Author:  gitsync.Identity{Name: "Somebody"},
		Paths:   []string{"content/a.md"},
		Message: title,
	}
}
