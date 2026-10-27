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
		RoleOwner: true, RoleImplementer: true, RoleEditor: true, RoleViewer: true,
	},
	ActionEditContent: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionEditMedia: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionPublish: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionGoBackToAVersion: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: true, RoleViewer: false,
	},
	ActionManagePeople: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionChangeSettings: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionManageWorkflows: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionSeeTechnicalDetail: {
		RoleOwner: true, RoleImplementer: true, RoleEditor: false, RoleViewer: false,
	},
	ActionManageBilling: {
		RoleOwner: true, RoleImplementer: false, RoleEditor: false, RoleViewer: false,
	},
	ActionDeleteProject: {
		RoleOwner: true, RoleImplementer: false, RoleEditor: false, RoleViewer: false,
	},
	ActionCreateProject: {
		RoleOwner: true, RoleImplementer: false, RoleEditor: false, RoleViewer: false,
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
	actor := holder(RoleOwner)

	elsewhere := Project(anOrganisation, "another-project")
	for _, action := range Actions {
		if Can(actor, action, elsewhere) {
			t.Errorf("an owner of one project may %s on another", action)
		}
	}
}

func TestOrganisationOwnerOwnsItsProjects(t *testing.T) {
	owner := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationOwner},
		Projects:      map[string]Role{},
	}

	role, held := EffectiveRole(owner, here())
	if !held || role != RoleOwner {
		t.Fatalf("EffectiveRole() = %q, %v, want owner", role, held)
	}

	for _, action := range Actions {
		if OverTheInstallation(action) {
			if Can(owner, action, here()) {
				t.Errorf("owning an organisation granted %s, which is the installation's", action)
			}
			continue
		}
		if !Can(owner, action, here()) {
			t.Errorf("an organisation's owner may not %s on its project", action)
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

func TestAnOrganisationOwnerKeepsOwnerEvenWithALesserProjectRole(t *testing.T) {
	actor := Actor{
		PersonID:      "person-1",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationOwner},
		Projects:      map[string]Role{aProject: RoleViewer},
	}

	if !Can(actor, ActionManageBilling, here()) {
		t.Error("an organisation's owner lost ownership by being added as a viewer")
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
	onAProject := 0
	for _, action := range Actions {
		if !OverTheInstallation(action) {
			onAProject++
		}
	}
	if cases[RoleOwner] != onAProject {
		t.Errorf("an owner may do %d of %d things on a project", cases[RoleOwner], onAProject)
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
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationOwner},
	}
	member := Actor{
		PersonID:      "person-2",
		Organisations: map[string]OrganisationRole{anOrganisation: OrganisationMember},
	}

	if !Can(owner, ActionCreateProject, Organisation(anOrganisation)) {
		t.Error("an organisation's owner may not create a project in it")
	}
	if Can(member, ActionCreateProject, Organisation(anOrganisation)) {
		t.Error("an organisation's member may create a project in it")
	}
	if Can(owner, ActionCreateProject, Organisation("another-organisation")) {
		t.Error("an owner of one organisation may create a project in another")
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
		t.Error("joining as an editor granted an owner's action")
	}
	if !Can(joined, ActionCreateOrganisation, Resource{}) {
		t.Error("joining a project took away administering")
	}
}

func TestWithoutAdministeringTheInstallationIsClosed(t *testing.T) {
	for _, who := range []Actor{
		Nobody(),
		{PersonID: "person-1"},
		holder(RoleOwner),
		{PersonID: "person-2", Organisations: map[string]OrganisationRole{anOrganisation: OrganisationOwner}},
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
		if Can(holder(RoleOwner), ActionCreateOrganisation, resource) {
			t.Error("a project owner was allowed because of the resource named")
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
