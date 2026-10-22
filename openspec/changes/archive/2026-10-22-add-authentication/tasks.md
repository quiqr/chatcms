## 1. The schema

- [x] 1.1 Add a migration creating people, sessions and single-use tokens,
  with the uniqueness and foreign keys each needs; verify it applies and rolls
  back with `TestMigrateUpDown`.
- [x] 1.2 Store addresses so that case does not create a second person; verify
  with `TestAddressesAreCaseInsensitive`.
- [x] 1.3 Add the queries and regenerate; verify `sqlc diff` is clean.

## 2. Passwords

- [x] 2.1 Add `internal/auth` with argon2id hashing whose encoded form carries
  its parameters; verify with `TestHashNamesItsParameters` and
  `TestTwoPeopleWithOnePasswordDifferInStorage`.
- [x] 2.2 Verify a password in constant time and accept a hash made with other
  parameters; verify with `TestVerifyAcceptsOtherParameters` and
  `TestVerifyRejectsAWrongPassword`.
- [x] 2.3 Refuse a password below the minimum length and one from the common
  list, and accept a long lower-case passphrase; verify with
  `TestPasswordRules`.
- [x] 2.4 Ship the common-password list in the binary; verify with
  `TestCommonListIsEmbedded`.

## 3. Sessions

- [x] 3.1 Issue a session as random bytes, storing only its hash; verify with
  `TestStoredSessionIsNotTheCookieValue`.
- [x] 3.2 Look up a session by the presented value, refusing unknown, expired
  and revoked ones; verify with `TestLookupRefusesEachBadSession`.
- [x] 3.3 Record when a session was last used without extending it; verify
  with `TestUsingASessionDoesNotExtendIt`.
- [x] 3.4 Give a remembered session a longer life than an ordinary one, in the
  stored row as well as the cookie; verify with `TestRememberedSessionsLast`.
- [x] 3.5 Revoke a session, and revoke every session a person has; verify with
  `TestRevokeOne` and `TestRevokeAllForAPerson`.

## 4. The identity seam

- [x] 4.1 Put checking who someone is behind one interface with a password
  implementation; verify with `TestPasswordIdentityChecks`.
- [x] 4.2 Make everything that reads a session work for a person with no
  password, and refuse a password sign-in for them; verify with
  `TestPersonWithNoPassword`.

## 5. The HTTP surface

- [x] 5.1 Replace the single anonymous actor with one taken from the session;
  verify with `TestRequestCarriesThePersonFromTheSession` and
  `TestRequestWithNoSessionSeesNobody`.
- [x] 5.2 Add the endpoint that requires a person and answers 401 in the JSON
  error shape otherwise; verify with `TestProtectedEndpointRefusesNobody`.
- [x] 5.3 Set the session cookie unreadable by scripts, same-site, and secure
  when the site is; verify with `TestSessionCookieIsProtected` for both an
  insecure and a secure base address.
- [x] 5.4 Add sign-in: new session every time, revoking whatever arrived;
  verify with `TestSignInIssuesANewSession` and
  `TestSignInRevokesTheSessionItArrivedWith`.
- [x] 5.5 Give the same answer for an unknown address and a wrong password,
  and spend comparable time on both; verify with
  `TestSignInSaysTheSameThingEitherWay` and `TestUnknownAddressStillHashes`.
- [x] 5.6 Add sign-out, revoking server-side and clearing the cookie, and
  succeeding when there is no session; verify with `TestSignOutRevokes` and
  `TestSignOutWithoutASession`.
- [x] 5.7 Add the double-submitted token for state-changing requests, not
  required for reads; verify with `TestChangeWithoutTheTokenIsRefused`,
  `TestChangeWithTheTokenProceeds` and `TestReadNeedsNoToken`.

## 6. Invitations and resets

- [x] 6.1 Issue a single-use expiring token, stored as a hash, for an
  invitation and for a reset; verify with `TestTokenIsStoredAsAHash`.
- [x] 6.2 Accept an invitation by setting the first password and spending the
  token; verify with `TestAcceptingAnInvitation`.
- [x] 6.3 Refuse a spent, an expired and a never-issued token, each in plain
  language and alike enough to reveal nothing; verify with
  `TestBadTokensAreRefusedAlike`.
- [x] 6.4 Answer a reset request the same way whether the address is known or
  not; verify with `TestResetSaysNothingAboutTheAddress`.
- [x] 6.5 Revoke every session when a reset is spent; verify with
  `TestResetEndsOtherSessions`.
- [x] 6.6 Record the link in the log and never in the response; verify with
  `TestLinkIsLoggedNotReturned`.

## 7. Rate limiting

- [x] 7.1 Limit per address and per caller; verify with
  `TestLimitPerAddress` and `TestLimitPerCaller`.
- [x] 7.2 Clear the count on a correct sign-in; verify with
  `TestSuccessClearsTheCount`.
- [x] 7.3 Say nothing about the address when refusing; verify with
  `TestRefusalSaysNothing`.

## 8. The screen

- [x] 8.1 Build the sign-in screen as the designs draw it, with the
  invitation hint; verify with `renders the sign-in form` and compare with
  `docs/design/reference/screenshots/desktop/login.png`.
- [x] 8.2 Keep it usable at 390 pixels; verify in the browser and record the
  result.
- [x] 8.3 Send the token with every change the screen makes; verify with
  `sends the request token`.

## 9. The gate and shipping

- [x] 9.1 Confirm `internal/auth` is well covered and `internal/httpapi` stays
  above its floor; verify with the coverage gate.
- [x] 9.2 Add bullets to `CHANGELOG.md` under `## [Unreleased]`.
- [x] 9.3 Run `nix flake check` and confirm every check passes.
