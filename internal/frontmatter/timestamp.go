package frontmatter

import (
	"regexp"
	"strconv"
	"time"
)

// The two forms js-yaml's timestamp resolver accepts, from the YAML 1.1
// specification. A scalar matching either is a moment in time, not a string,
// which is why saving a page rewrites a bare date as a full timestamp: the
// reference implementation does exactly the same.
var (
	timestampDateOnly = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	timestampFull     = regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})` +
		`(?:[Tt]|[ \t]+)(\d{1,2}):(\d{2}):(\d{2})(?:\.(\d*))?` +
		`(?:[ \t]*(?:(Z)|([-+])(\d{1,2})(?::(\d{2}))?))?$`)
)

// asTimestamp reports whether a scalar is a moment in time, and which.
func asTimestamp(value string) (time.Time, bool) {
	if match := timestampDateOnly.FindStringSubmatch(value); match != nil {
		year, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
	}

	match := timestampFull.FindStringSubmatch(value)
	if match == nil {
		return time.Time{}, false
	}

	year, _ := strconv.Atoi(match[1])
	month, _ := strconv.Atoi(match[2])
	day, _ := strconv.Atoi(match[3])
	hour, _ := strconv.Atoi(match[4])
	minute, _ := strconv.Atoi(match[5])
	second, _ := strconv.Atoi(match[6])

	nanoseconds := 0
	if fraction := match[7]; fraction != "" {
		// A fraction is read to nanosecond resolution, padded or cut.
		digits := (fraction + "000000000")[:9]
		nanoseconds, _ = strconv.Atoi(digits)
	}

	location := time.UTC
	if match[9] != "" {
		offsetHours, _ := strconv.Atoi(match[10])
		offsetMinutes := 0
		if match[11] != "" {
			offsetMinutes, _ = strconv.Atoi(match[11])
		}
		seconds := offsetHours*3600 + offsetMinutes*60
		if match[9] == "-" {
			seconds = -seconds
		}
		location = time.FixedZone("", seconds)
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, location).UTC(), true
}

// looksLikeTimestamp is used when deciding whether a string has to be quoted,
// so that a string someone typed that happens to read as a date is written
// back as a string rather than silently becoming a date.
func looksLikeTimestamp(value string) bool {
	_, ok := asTimestamp(value)
	return ok
}
