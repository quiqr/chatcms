package auth

import (
	_ "embed"
	"strings"
	"sync"
)

// commonPasswords ships in the binary rather than being fetched, because a
// check that depends on a network call is a check that is sometimes skipped.
//
//go:embed common-passwords.txt
var commonPasswords string

var commonSet = sync.OnceValue(func() map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(commonPasswords, "\n") {
		word := strings.TrimSpace(line)
		if word != "" {
			out[strings.ToLower(word)] = true
		}
	}
	return out
})

func isCommonPassword(password string) bool {
	return commonSet()[strings.ToLower(password)]
}

// CommonPasswordCount is how many passwords the list holds, which the tests
// use to prove it was embedded rather than empty.
func CommonPasswordCount() int {
	return len(commonSet())
}
