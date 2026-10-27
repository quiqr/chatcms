# Tasks

## 1. The server knows who is asking and where

- [x] 1.1 One answer carrying your name, your organisations, and this one
- [x] 1.2 An organisation you are not in is nothing there, never a redirect
- [x] 1.3 Renaming, which needs administering
- [x] 1.4 The people of an organisation, which its owner manages
- [x] 1.5 Refuse a change that would leave it with no owner

## 2. The address carries the organisation

- [x] 2.1 Routes take it, in App.tsx and in every link
- [x] 2.2 The eight hardcoded slugs go
- [x] 2.3 The shells read it from a context asked once

## 3. The screens

- [x] 3.1 The switcher lists what you may switch to and changes the address
- [x] 3.2 Starting one, offered only to an administrator
- [x] 3.3 The organisation screen: name, address, people, renaming
- [x] 3.4 Belonging to none says so and says what to do
- [x] 3.5 At 1440 px and at 390 px

## 4. Proving it

- [x] 4.1 An integration test that one organisation cannot read another's
- [x] 4.2 A unit test for each refusal on the screens
- [x] 4.3 An acceptance test that switches between two and sees each one's own
- [x] 4.4 A check that no file under web/src names an organisation
