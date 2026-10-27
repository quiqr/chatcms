package history

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/publish"
)

// Entry is one version of the site as a person sees it in the timeline.
type Entry struct {
	// Version is the identifier the engine gave it. It is never shown; it is
	// what a request to go back names.
	Version string    `json:"version"`
	Title   string    `json:"title"`
	By      string    `json:"by,omitempty"`
	At      time.Time `json:"at"`
	Changes int       `json:"changes"`
	// Live says this is the version the published site holds.
	Live bool `json:"live"`
	// Note is what somebody wrote when they published it.
	Note string `json:"note,omitempty"`
}

// Group is a stretch of time the timeline is divided into, in the words a
// person would use for it.
type Group struct {
	Period  string  `json:"period"`
	Entries []Entry `json:"entries"`
}

// Timeline is the history of a site: what was published when, plus the work
// that has not been published yet.
type Timeline struct {
	Unpublished int     `json:"unpublished"`
	Groups      []Group `json:"groups"`
}

// Build turns versions and publish notes into the timeline. The note a person
// wrote wins over anything generated, because they wrote it to be read.
func Build(versions []gitsync.Version, notes map[string]string, live string, unpublished int, now time.Time) Timeline {
	timeline := Timeline{Unpublished: unpublished, Groups: []Group{}}

	ordered := append([]gitsync.Version{}, versions...)
	sort.SliceStable(ordered, func(a, b int) bool { return ordered[a].Time.After(ordered[b].Time) })

	for _, version := range ordered {
		entry := Entry{
			Version: version.ID,
			Title:   titleOf(version, notes[version.ID]),
			By:      version.Author.Name,
			At:      version.Time,
			Changes: len(version.Paths),
			Live:    version.ID == live,
			Note:    notes[version.ID],
		}
		timeline.add(periodOf(version.Time, now), entry)
	}
	return timeline
}

func (t *Timeline) add(period string, entry Entry) {
	for index := range t.Groups {
		if t.Groups[index].Period == period {
			t.Groups[index].Entries = append(t.Groups[index].Entries, entry)
			return
		}
	}
	t.Groups = append(t.Groups, Group{Period: period, Entries: []Entry{entry}})
}

// titleOf is what a version is called. A publish note is what somebody chose
// to say about it; without one, the version's own message is used, and a
// message this product wrote itself is turned into something readable.
func titleOf(version gitsync.Version, note string) string {
	if note != "" {
		return note
	}
	if version.Message != "" && !looksTechnical(version.Message) {
		return version.Message
	}
	return summarise(version.Paths)
}

func looksTechnical(message string) bool {
	lowered := strings.ToLower(message)
	for _, word := range []string{"merge", "rebase", "revert", "cherry-pick", "squash", "fixup"} {
		if strings.HasPrefix(lowered, word) {
			return true
		}
	}
	return false
}

// summarise says what a version did when nobody said it in words.
func summarise(paths []string) string {
	if len(paths) == 0 {
		return "A small change"
	}

	pages, images, other := 0, 0, 0
	for _, path := range paths {
		switch {
		case publish.IsImage(path):
			images++
		case publish.IsContent(path):
			pages++
		default:
			other++
		}
	}

	said := []string{}
	said = appendCount(said, pages, "page", "pages")
	said = appendCount(said, images, "image", "images")
	said = appendCount(said, other, "other file", "other files")
	if len(said) == 0 {
		return "A small change"
	}
	return capitalise(strings.Join(said, " and ") + " changed")
}

func appendCount(said []string, count int, one, many string) []string {
	if count == 0 {
		return said
	}
	if count == 1 {
		return append(said, "1 "+one)
	}
	return append(said, fmt.Sprintf("%d %s", count, many))
}

func capitalise(said string) string {
	if said == "" {
		return said
	}
	return strings.ToUpper(said[:1]) + said[1:]
}

// periodOf is the stretch of time a version falls in, said the way a person
// would say it rather than as a date.
func periodOf(at, now time.Time) string {
	if at.IsZero() {
		return "Earlier"
	}

	// Both are read in one zone before their days are compared. A version
	// carries the offset it was written in, so a change made ten minutes ago
	// in a different offset would otherwise be filed under Yesterday.
	here := at.In(now.Location())
	today := startOfDay(now)
	switch day := startOfDay(here); {
	case !day.Before(today):
		return "Today"
	case !day.Before(today.AddDate(0, 0, -1)):
		return "Yesterday"
	case !day.Before(today.AddDate(0, 0, -7)):
		return "Earlier this week"
	case here.Year() == now.Year() && here.Month() == now.Month():
		return "Earlier this month"
	case here.Year() == now.Year():
		return here.Format("January")
	default:
		return here.Format("January 2006")
	}
}

func startOfDay(at time.Time) time.Time {
	year, month, day := at.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, at.Location())
}
