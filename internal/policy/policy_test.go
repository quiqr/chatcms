package policy

import (
	"testing"
)

const (
	anOrganisation = "org-1"
	aProject       = "project-1"
)

func holder(role Role) Actor {
	return Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationMember},
		Projects:      map[string]Role{aProject: role},
	}
}

func here() Resource {
	return Project(anOrganisation, aProject)
}

// expected is docs/design/DESIGN.md section 6, written out rather than worked
// out, so that a mistake in the implementation's table does not become the
// test's expectation as well.
var expected = map[Action]map[Role]bool{
	ActionReadContent: {
		RoleImplementer: true, RoleEditor: true, RoleViewer: true,
	},
	ActionEditContent: {
		RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionEditMedia: {
		RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionPublish: {
		RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionGoBackToAVersion: {
		RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionManagePeople: {
		RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionChangeSettings: {
		RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionManageWorkflows: {
		RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionSeeTechnicalDetail: {
		RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
}

func TestEveryRoleAgainstEveryAction(t *testing.T) {
	for _, action := range Actions {
		// An action held over the installation is not in the table of roles:
		// administering is not a role, and no role grants it.
		if OverTheInstallation(action) {
			if _, listed := expected[action]; listed {
				t.Errorf("%s is held over the installation and should not be in the role table", action)
			}
			continue
		}

		// Nor is an action held over the organisation: no project role
		// carries one, however large it is.
		if OverTheOrganisation(action) {
			if _, listed := expected[action]; listed {
				t.Errorf("%s is held over the organisation and should not be in the role table", action)
			}
			continue
		}

		answers, listed := expected[action]
		if !listed {
			t.Errorf("%s is an action with no expectation written for it", action)
			continue
		}

		for _, role := range Roles {
			want, written := answers[role]
			if !written {
				t.Errorf("%s has no expectation for %s", action, role)
				continue
			}

			t.Run(string(action)+"/"+string(role), func(t *testing.T) {
				if got := Can(holder(role), action, here()); got != want {
					t.Errorf("Can(%s, %s) = %v, want %v", role, action, got, want)
				}
			})
		}
	}

	for action := range expected {
		found := false
		for _, known := range Actions {
			if known == action {
				found = true
			}
		}
		if !found {
			t.Errorf("there is an expectation for %s, which is not an action", action)
		}
	}
}

func TestNobodyMayNothing(t *testing.T) {
	for _, action := range Actions {
		if Can(Nobody(), action, here()) {
			t.Errorf("nobody may %s", action)
		}
		if Can(Nobody(), action, Organisation(anOrganisation)) {
			t.Errorf("nobody may %s on an organisation", action)
		}
	}
}

func TestNoRoleMeansNo(t *testing.T) {
	stranger := Actor{PersonID: "person-2"}

	for _, action := range Actions {
		if Can(stranger, action, here()) {
			t.Errorf("someone with no role may %s", action)
		}
	}

	if _, held := EffectiveRole(stranger, here()); held {
		t.Error("someone with no role holds one")
	}
}

func TestAnotherProjectIsNotThisOne(t *testing.T) {
	actor := holder(RoleImplementer)

	elsewhere := Project(anOrganisation, "another-project")
	for _, action := range Actions {
		if Can(actor, action, elsewhere) {
			t.Errorf("an implementer of one project may %s on another", action)
		}
	}
}

func TestOrganisationAdminReachesItsProjects(t *testing.T) {
	admin := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin},
		Projects:      map[string]Role{},
	}

	// Implementer, because there is no owner of a project to raise to. What
	// an organisation admin holds beyond that is held over the organisation.
	role, held := EffectiveRole(admin, here())
	if !held || role != RoleImplementer {
		t.Fatalf("EffectiveRole() = %q, %v, want implementer", role, held)
	}

	for _, action := range Actions {
		if OverTheInstallation(action) {
			if Can(admin, action, here()) {
				t.Errorf("administering an organisation granted %s, which is the installation's", action)
			}
			continue
		}
		if !Can(admin, action, here()) {
			t.Errorf("an organisation's admin may not %s on its project", action)
		}
	}
}

func TestOrganisationMemberDoesNot(t *testing.T) {
	member := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationMember},
		Projects:      map[string]Role{},
	}

	for _, action := range Actions {
		if Can(member, action, here()) {
			t.Errorf("an organisation's member with no project role may %s", action)
		}
	}
}

func TestAProjectRoleIsNotRaisedByOrganisationMembership(t *testing.T) {
	editor := holder(RoleEditor)

	if Can(editor, ActionManagePeople, here()) {
		t.Error("an editor who is an organisation member may manage people")
	}
}

func TestAnOrganisationAdminKeepsItEvenWithALesserProjectRole(t *testing.T) {
	actor := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin},
		Projects:      map[string]Role{aProject: RoleViewer},
	}

	if !Can(actor, ActionManageBilling, here()) {
		t.Error("an organisation's admin lost it by being added as a viewer")
	}
	if !Can(actor, ActionEditContent, here()) {
		t.Error("being added as a viewer took away what administering reaches")
	}
}

// TestARoleBelongsToAPlace holds the sentence the hierarchy board makes: a
// role is held over a place, not by a person. Somebody can administer one
// organisation and be an editor on a project in another, and neither reaches
// into the other.
func TestARoleBelongsToAPlace(t *testing.T) {
	const elsewhere = "org-2"
	const theirProject = "project-2"

	actor := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin, elsewhere: OrganisationMember},
		Projects:      map[string]Role{theirProject: RoleEditor},
	}

	if !Can(actor, ActionDeleteProject, here()) {
		t.Error("the admin of an organisation may not delete a project in it")
	}
	if Can(actor, ActionDeleteProject, Project(elsewhere, theirProject)) {
		t.Error("administering one organisation deleted a project in another")
	}
	if !Can(actor, ActionEditContent, Project(elsewhere, theirProject)) {
		t.Error("an editor elsewhere may not edit there")
	}
	if Can(actor, ActionManagePeople, Project(elsewhere, theirProject)) {
		t.Error("an editor elsewhere manages people there")
	}
}

func TestResourceCarriesWhatIsActedOn(t *testing.T) {
	actor := holder(RoleEditor)

	resource := here()
	resource.Within = "content/about.md"

	if !Can(actor, ActionEditContent, resource) {
		t.Error("naming something inside the project changed the answer")
	}
	if resource.ProjectID != aProject || resource.OrganisationID != anOrganisation {
		t.Error("the resource lost what it names")
	}
}

func TestAllowedListsWhatMayBeDone(t *testing.T) {
	cases := map[Role]int{}
	for _, role := range Roles {
		cases[role] = len(Allowed(holder(role), here()))
	}

	if cases[RoleViewer] != 1 {
		t.Errorf("a viewer may do %d things, want only reading", cases[RoleViewer])
	}
	withARole := 0
	for _, action := range Actions {
		if !OverTheInstallation(action) && !OverTheOrganisation(action) {
			withARole++
		}
	}
	if cases[RoleImplementer] != withARole {
		t.Errorf("an implementer may do %d of %d things a role carries", cases[RoleImplementer], withARole)
	}

	// An organisation's admin is offered what the organisation carries as
	// well, because a screen asking what it may offer on a project has to
	// know whether Delete belongs on it.
	admin := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin},
	}
	onAProject := 0
	for _, action := range Actions {
		if !OverTheInstallation(action) {
			onAProject++
		}
	}
	if len(Allowed(admin, here())) != onAProject {
		t.Errorf("an organisation's admin is offered %d of %d things on a project",
			len(Allowed(admin, here())), onAProject)
	}

	if cases[RoleEditor] >= cases[RoleImplementer] {
		t.Error("an editor may do at least as much as an implementer")
	}
	if len(Allowed(Nobody(), here())) != 0 {
		t.Error("nobody may do something")
	}
}

func TestOrganisationLevelActions(t *testing.T) {
	owner := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin},
	}
	member := Actor{
		PersonID:      "person-2",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationMember},
	}

	for _, action := range Actions {
		if !OverTheOrganisation(action) {
			continue
		}
		if !Can(owner, action, Organisation(anOrganisation)) {
			t.Errorf("an organisation's admin may not %s in it", action)
		}
		if Can(member, action, Organisation(anOrganisation)) {
			t.Errorf("an organisation's member may %s in it", action)
		}
		if Can(owner, action, Organisation("another-organisation")) {
			t.Errorf("the admin of one organisation may %s in another", action)
		}
		if Can(holder(RoleImplementer), action, here()) {
			t.Errorf("a project's implementer may %s, which belongs to the organisation", action)
		}
	}
}

// TestSettingLimitsIsThePlatformAdmins holds the top of the hierarchy: a
// limit is set from above the organisation it binds, so no organisation admin
// can lift their own.
func TestSettingLimitsIsThePlatformAdmins(t *testing.T) {
	if !OverTheInstallation(ActionSetLimits) {
		t.Fatal("setting limits is not held over the installation")
	}

	admin := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin},
	}
	if Can(admin, ActionSetLimits, Organisation(anOrganisation)) {
		t.Error("an organisation's admin may set its own limits")
	}
	if !Can(administrator(), ActionSetLimits, Resource{}) {
		t.Error("a platform admin may not set limits")
	}
}

// administrator is somebody who administers the installation and holds
// nothing anywhere else, which is the case the whole design turns on.
func administrator() Actor {
	return Actor{PersonID: "person-admin", Administers: true}
}

func TestAnAdministratorMayShapeOrganisations(t *testing.T) {
	for _, action := range Actions {
		if !OverTheInstallation(action) {
			continue
		}
		if !Can(administrator(), action, Resource{}) {
			t.Errorf("an administrator may not %s", action)
		}
	}
}

func TestAdministeringIsNotEditing(t *testing.T) {
	// The whole point: whoever runs the service decides who may reach a
	// customer's site without being able to read or change it themselves.
	for _, action := range Actions {
		if OverTheInstallation(action) {
			continue
		}
		if Can(administrator(), action, here()) {
			t.Errorf("administering granted %s on a customer's project", action)
		}
	}

	if _, held := EffectiveRole(administrator(), here()); held {
		t.Error("an administrator holds a role on a project they are not in")
	}
	if len(Allowed(administrator(), here())) != 0 {
		t.Error("an administrator is allowed something on a project they are not in")
	}
}

func TestAnAdministratorWhoJoinsHoldsWhatTheJoiningGrants(t *testing.T) {
	joined := administrator()
	joined.Projects = map[string]Role{aProject: RoleEditor}

	if !Can(joined, ActionEditContent, here()) {
		t.Error("an administrator who joined as an editor may not edit")
	}
	if Can(joined, ActionDeleteProject, here()) {
		t.Error("joining as an editor granted what the organisation holds")
	}
	if !Can(joined, ActionCreateOrganisation, Resource{}) {
		t.Error("joining a project took away administering")
	}
}

func TestWithoutAdministeringTheInstallationIsClosed(t *testing.T) {
	for _, who := range []Actor{
		Nobody(),
		{PersonID: "person-1"},
		holder(RoleImplementer),
		{PersonID: "person-2", Organisations: map[string]OrganisationRole{anOrganisation: OrganisationAdmin}},
	} {
		for _, action := range Actions {
			if !OverTheInstallation(action) {
				continue
			}
			if Can(who, action, Resource{}) {
				t.Errorf("%q was allowed %s without administering", who.PersonID, action)
			}
		}
	}
}

func TestTheResourceSaysNothingAboutInstallationActions(t *testing.T) {
	// Asking about an installation action while naming somebody else's
	// project must not change the answer, in either direction.
	for _, resource := range []Resource{{}, here(), Organisation("another-organisation")} {
		if !Can(administrator(), ActionCreateOrganisation, resource) {
			t.Error("an administrator was refused because of the resource named")
		}
		if Can(holder(RoleImplementer), ActionCreateOrganisation, resource) {
			t.Error("a project implementer was allowed because of the resource named")
		}
	}
}

func TestSomebodyWhoAdministersCannotBeActedAs(t *testing.T) {
	// Acting as an administrator would be a way of becoming one, which is
	// the whole reason this is decided here rather than beside the act.
	if MayBeActedAs(Actor{PersonID: "a-person", Administers: true}) {
		t.Error("an administrator may be acted as")
	}
	if !MayBeActedAs(Actor{PersonID: "a-person"}) {
		t.Error("an ordinary person may not be acted as")
	}
}
