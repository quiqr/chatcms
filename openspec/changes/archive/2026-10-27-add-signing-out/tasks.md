# Tasks

## 1. Signing out everywhere

- [x] 1.1 An endpoint that ends every session this account has
- [x] 1.2 It ends the one that asked as well
- [x] 1.3 It touches nobody else's

## 2. The person menu

- [x] 2.1 One component: who you are, your account, signing out
- [x] 2.2 Used by all four shells
- [x] 2.3 A test that walks all four, so none can lack it again
- [x] 2.4 On a phone, from the person at the top

## 3. What happens afterwards

- [x] 3.1 Signing out lands on the sign-in screen saying so
- [x] 3.2 A session that ran out does not say it

## 4. Proving it

- [x] 4.1 Unit tests on the menu and the shells
- [x] 4.2 An integration test that every session really ends
- [x] 4.3 An acceptance test: sign in, sign out, cannot get back
