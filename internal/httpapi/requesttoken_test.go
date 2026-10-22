package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func nowPlusAnHour() time.Time {
	return time.Now().Add(time.Hour)
}

func guarded() http.Handler {
	return RequireRequestToken(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"did": "change something"})
	}))
}

func TestChangeWithoutTheTokenIsRefused(t *testing.T) {
	cases := map[string]func(*http.Request){
		"no cookie and no header": func(*http.Request) {},
		"a cookie but no header": func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
		},
		"a header but no cookie": func(r *http.Request) {
			r.Header.Set(RequestTokenHeader, "a-token")
		},
		"they do not match": func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
			r.Header.Set(RequestTokenHeader, "another-token")
		},
	}

	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/anything", nil)
			prepare(request)
			recorder := httptest.NewRecorder()

			guarded().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", recorder.Code)
			}
			if recorder.Body.String() == "" {
				t.Error("the refusal has no body")
			}
			apiErr := decodeError(t, recorder.Body.Bytes())
			if apiErr.Code != CodeForbidden {
				t.Errorf("code = %q, want %q", apiErr.Code, CodeForbidden)
			}
			if strings.Contains(recorder.Body.String(), "change something") {
				t.Error("the handler ran although the request was refused")
			}
		})
	}
}

func TestChangeWithTheTokenProceeds(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/anything", nil)
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
	request.Header.Set(RequestTokenHeader, "a-token")

	recorder := httptest.NewRecorder()
	guarded().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", recorder.Code)
	}
}

func TestReadNeedsNoToken(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		recorder := httptest.NewRecorder()
		guarded().ServeHTTP(recorder, httptest.NewRequest(method, "/api/anything", nil))

		if recorder.Code != http.StatusOK {
			t.Errorf("%s status = %d, want 200", method, recorder.Code)
		}
	}
}

func TestABrowserWithNoTokenIsGivenOne(t *testing.T) {
	recorder := httptest.NewRecorder()
	guarded().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	var token *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == RequestTokenCookie {
			token = cookie
		}
	}

	if token == nil {
		t.Fatal("no request token was given to a browser that had none")
	}
	if token.Value == "" {
		t.Error("the request token is empty")
	}
	if token.HttpOnly {
		t.Error("the request token cannot be read by the page, so the page cannot repeat it")
	}
}

func TestAnExistingTokenIsNotReplaced(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "already-have-one"})

	recorder := httptest.NewRecorder()
	guarded().ServeHTTP(recorder, request)

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == RequestTokenCookie {
			t.Errorf("the token was replaced with %q", cookie.Value)
		}
	}
}

func TestSessionCookieIsProtected(t *testing.T) {
	for _, secure := range []bool{false, true} {
		recorder := httptest.NewRecorder()
		SetSessionCookie(recorder, "a-session", nowPlusAnHour(), secure)

		cookies := recorder.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("got %d cookies, want 1", len(cookies))
		}
		cookie := cookies[0]

		if cookie.Name != SessionCookie {
			t.Errorf("name = %q, want %q", cookie.Name, SessionCookie)
		}
		if !cookie.HttpOnly {
			t.Error("the session cookie can be read by page scripts")
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("SameSite = %v, want Lax", cookie.SameSite)
		}
		if cookie.Secure != secure {
			t.Errorf("Secure = %v for secure=%v", cookie.Secure, secure)
		}
		if cookie.Path != "/" {
			t.Errorf("Path = %q, want /", cookie.Path)
		}
	}
}

func TestClearSessionCookieExpiresIt(t *testing.T) {
	recorder := httptest.NewRecorder()
	ClearSessionCookie(recorder, true)

	cookie := recorder.Result().Cookies()[0]
	if cookie.Value != "" || cookie.MaxAge >= 0 {
		t.Errorf("cookie = %+v, want it emptied and expired", cookie)
	}
}
