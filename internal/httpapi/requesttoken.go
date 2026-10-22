package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

// safeMethod reports whether a request only reads. A request that changes
// nothing cannot be abused through another site, so it needs no token.
func safeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// RequireRequestToken refuses a state-changing request that does not repeat
// the value of the token cookie in a header. Another site can make a browser
// send a cookie, but cannot read one to repeat it.
func RequireRequestToken(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ensureRequestToken(w, r, secure)

			if safeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(RequestTokenCookie)
			presented := r.Header.Get(RequestTokenHeader)

			if err != nil || cookie.Value == "" || presented == "" ||
				subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(presented)) != 1 {
				WriteError(w, http.StatusForbidden, CodeForbidden,
					"This request could not be confirmed. Reload the page and try again.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ensureRequestToken gives the browser a token if it does not have one, so
// that the first state-changing request of a session can carry it.
func ensureRequestToken(w http.ResponseWriter, r *http.Request, secure bool) {
	if cookie, err := r.Cookie(RequestTokenCookie); err == nil && cookie.Value != "" {
		return
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RequestTokenCookie,
		Value:    base64.RawURLEncoding.EncodeToString(raw),
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
