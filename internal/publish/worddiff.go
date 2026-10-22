package publish

import (
	"strings"
	"unicode"
)

// Part says how a run of words relates to the published text.
type Part string

const (
	Kept  Part = "kept"
	Gone  Part = "gone"
	Added Part = "added"
)

// Word is a run of text and what happened to it, which is what the designs
// show struck through in red or highlighted in green.
type Word struct {
	Part Part   `json:"part"`
	Text string `json:"text"`
}

const maxWordsCompared = 4000

// WordDiff compares two pieces of text word by word. Comparing by word rather
// than by character is what makes the result readable: a person wants to see
// which words changed, not which letters.
func WordDiff(was, now string) []Word {
	if was == now {
		return nil
	}

	before, after := split(was), split(now)
	if len(before) > maxWordsCompared || len(after) > maxWordsCompared {
		return wholesale(was, now)
	}

	head := 0
	for head < len(before) && head < len(after) && before[head] == after[head] {
		head++
	}
	tail := 0
	for tail < len(before)-head && tail < len(after)-head &&
		before[len(before)-1-tail] == after[len(after)-1-tail] {
		tail++
	}

	middleBefore := before[head : len(before)-tail]
	middleAfter := after[head : len(after)-tail]

	parts := []Word{}
	parts = add(parts, Kept, before[:head])
	for _, run := range longestCommon(middleBefore, middleAfter) {
		parts = add(parts, run.part, run.words)
	}
	parts = add(parts, Kept, before[len(before)-tail:])
	return join(parts)
}

func wholesale(was, now string) []Word {
	parts := []Word{}
	if was != "" {
		parts = append(parts, Word{Part: Gone, Text: was})
	}
	if now != "" {
		parts = append(parts, Word{Part: Added, Text: now})
	}
	return parts
}

type run struct {
	part  Part
	words []string
}

// longestCommon is the usual longest-common-subsequence walk. The texts
// compared here are a page's fields, so the quadratic table is small; the
// word cap above is what keeps it that way.
func longestCommon(before, after []string) []run {
	lengths := make([][]int, len(before)+1)
	for i := range lengths {
		lengths[i] = make([]int, len(after)+1)
	}
	for i := len(before) - 1; i >= 0; i-- {
		for j := len(after) - 1; j >= 0; j-- {
			if before[i] == after[j] {
				lengths[i][j] = lengths[i+1][j+1] + 1
			} else if lengths[i+1][j] >= lengths[i][j+1] {
				lengths[i][j] = lengths[i+1][j]
			} else {
				lengths[i][j] = lengths[i][j+1]
			}
		}
	}

	runs := []run{}
	i, j := 0, 0
	for i < len(before) && j < len(after) {
		switch {
		case before[i] == after[j]:
			runs = append(runs, run{Kept, []string{before[i]}})
			i, j = i+1, j+1
		case lengths[i+1][j] >= lengths[i][j+1]:
			runs = append(runs, run{Gone, []string{before[i]}})
			i++
		default:
			runs = append(runs, run{Added, []string{after[j]}})
			j++
		}
	}
	for ; i < len(before); i++ {
		runs = append(runs, run{Gone, []string{before[i]}})
	}
	for ; j < len(after); j++ {
		runs = append(runs, run{Added, []string{after[j]}})
	}
	return runs
}

func add(parts []Word, part Part, words []string) []Word {
	if len(words) == 0 {
		return parts
	}
	return append(parts, Word{Part: part, Text: strings.Join(words, "")})
}

// join puts neighbouring runs of the same sort together, so the result reads
// as phrases rather than as one entry per word.
func join(parts []Word) []Word {
	out := []Word{}
	for _, part := range parts {
		if part.Text == "" {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Part == part.Part {
			out[len(out)-1].Text += part.Text
			continue
		}
		out = append(out, part)
	}
	return out
}

// split cuts text into words with the space that follows them attached, so
// putting the pieces back together gives the text again.
func split(text string) []string {
	words := []string{}
	current := strings.Builder{}
	sawSpace := false

	for _, letter := range text {
		space := unicode.IsSpace(letter)
		if !space && sawSpace {
			words = append(words, current.String())
			current.Reset()
		}
		current.WriteRune(letter)
		sawSpace = space
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}
