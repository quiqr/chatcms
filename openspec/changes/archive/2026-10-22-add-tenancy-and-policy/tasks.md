## 1. The schema

- [x] 1.1 Add a migration creating organisations, projects, organisation
  memberships and project memberships, with their uniqueness and keys; verify
  it applies and rolls back.
- [x] 1.2 Give organisations and projects an address of their own that is
  unique where it has to be; verify with `TestSlugsAreUnique`.
- [x] 1.3 Add the queries and regenerate; verify `sqlc diff` is clean.

## 2. The policy

- [x] 2.1 Add `internal/policy` with the actions, the roles and a pure `Can`;
  verify it compiles with no import of the database.
- [x] 2.2 Write the role table from `docs/design/DESIGN.md` section 6; verify
  with `TestEveryRoleAgainstEveryAction`, whose expectations are written out
  separately from the implementation's table.
- [x] 2.3 Refuse everything to nobody; verify with `TestNobodyMayNothing`.
- [x] 2.4 Refuse everything on a resource the actor has no role on; verify
  with `TestNoRoleMeansNo`.
- [x] 2.5 Treat an organisation's owner as an owner of its projects; verify
  with `TestOrganisationOwnerOwnsItsProjects` and
  `TestOrganisationMemberDoesNot`.
- [x] 2.6 Take a resource that may name something within a project; verify
  with `TestResourceCarriesWhatIsActedOn`.
- [x] 2.7 Add `TestOnlyPolicyNamesARole`, walking the sources with the
  exceptions listed explicitly; verify by adding a temporary offender.

## 3. Organisations and projects

- [x] 3.1 Add `internal/tenancy` with creating an organisation whose creator
  owns it; verify with `TestCreatingAnOrganisation`.
- [x] 3.2 Create a project in an organisation; verify with
  `TestCreatingAProject`.
- [x] 3.3 List only what the asker may reach; verify with
  `TestListingShowsOnlyWhatIsReachable`.
- [x] 3.4 Load a person's memberships in one query; verify with
  `TestMembershipsLoadInOneQuery`.
- [x] 3.5 Refuse a removal that would leave an organisation or a project with
  nobody who can administer it; verify with `TestLastOwnerCannotBeRemoved`.
- [x] 3.6 Keep two organisations separate; verify with
  `TestOrganisationsAreSeparate`.

## 4. People on a project

- [x] 4.1 Invite by address and role, creating the person when new and issuing
  the link; verify with `TestInvitingSomeoneNew`.
- [x] 4.2 Give the role to a person who already exists, making no second
  person; verify with `TestInvitingSomeoneKnown`.
- [x] 4.3 Change rather than double a role when someone is invited twice;
  verify with `TestInvitingTwiceChangesTheRole`.
- [x] 4.4 Change a role and remove a person; verify with `TestChangingARole`
  and `TestRemovingAPerson`.
- [x] 4.5 List the people on a project with their roles and whether they are
  still to accept; verify with `TestListingPeople`.

## 5. The HTTP surface

- [x] 5.1 Add the endpoints for organisations, projects, people and
  invitations; verify each with its own test.
- [x] 5.2 Refuse what the actor may not do with 403, changing nothing; verify
  with `TestForbiddenChangesNothing`.
- [x] 5.3 Answer the same for a resource in another organisation as for one
  that does not exist; verify with `TestSomebodyElsesProjectLooksMissing`.
- [x] 5.4 Answer 401 rather than 403 for nobody; verify with
  `TestNobodyGetsUnauthorized`.
- [x] 5.5 Carry what this person may do in a project's own view; verify with
  `TestProjectViewCarriesPermissions`.
- [x] 5.6 Prove a viewer cannot change anything by calling every mutating
  endpoint; verify with `TestAViewerCannotChangeAnything`.

## 6. The screens

- [x] 6.1 Build the project list with the organisation switcher, reading real
  projects; verify with `lists the projects` and compare with
  `docs/design/reference/screenshots/desktop/projects.png`.
- [x] 6.2 Build people and roles with the invite row, the table and the role
  explanations; verify with `lists the people` and compare with
  `people.png`.
- [x] 6.3 Leave out what the person may not do rather than disabling it;
  verify with `an editor sees no invite row`.
- [x] 6.4 Check both at 390 pixels in a browser and record the result.

## 7. Acceptance and shipping

- [x] 7.1 Prove acceptance test 1 at the API level: create an organisation,
  create a project, invite someone as an editor, and have them sign in with
  the link; verify with `TestAcceptanceCreateOrganisationAndInviteAnEditor`.
- [x] 7.2 Confirm `internal/policy` is at or above the core floor, so the gate
  reports no absent core package; verify with the coverage gate.
- [x] 7.3 Add bullets to `CHANGELOG.md` under `## [Unreleased]`.
- [x] 7.4 Run `nix flake check` and confirm every check passes.
