# Tasks

## 1. The policy learns a level above the organisation

- [x] 1.1 An actor carries whether they administer the installation
- [x] 1.2 Actions held over the installation, answered from that alone
- [x] 1.3 Administering grants nothing on a project, with a test that says so
- [x] 1.4 The no-roles-outside-policy test covers the new flag

## 2. The database

- [x] 2.1 A migration: the flag on a person, and the record of what was done
- [x] 2.2 Queries: promote, demote, count, list, record, read the record
- [x] 2.3 Refuse the last removal in SQL as well as above it

## 3. internal/admin

- [x] 3.1 Make an organisation, with an owner named, recorded
- [x] 3.2 Rename and remove an organisation, recorded
- [x] 3.3 Put somebody into an organisation and take them out, recorded
- [x] 3.4 Appoint and remove administrators, with the last one kept
- [x] 3.5 List every organisation, and read the record

## 4. Appointing the first one

- [x] 4.1 A setting, as a value or from a file
- [x] 4.2 Promote on start-up, idempotently, saying what it did
- [x] 4.3 Start anyway when there is no such person, and say why
- [x] 4.4 The NixOS module and scripts/dev.sh pass it

## 5. Over the wire and on the screen

- [x] 5.1 Endpoints under /api/administration, each asking the policy
- [x] 5.2 Creating an organisation refuses anybody else
- [x] 5.3 The administration screens: organisations, people, the record
- [x] 5.4 The entry appears only for an administrator
- [x] 5.5 At 1440 px and at 390 px

## 6. Proving it

- [x] 6.1 Unit tests on the policy
- [x] 6.2 Integration tests against a real database for every refusal
- [x] 6.3 An administrator cannot read a customer's content
- [x] 6.4 Acceptance test 1 runs as an administrator
- [x] 6.5 cmd/seed makes one
