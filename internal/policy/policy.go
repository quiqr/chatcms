// Package policy decides who may do what. It is the only place in this
// codebase where a role is compared, which a test over the sources enforces.
// See docs/ARCHITECTURE.md and openspec/specs/tenancy/permissions.
//
// Nothing here reaches for a database. The caller loads the memberships; this
// package decides. That is what makes the whole table testable with the
// answers written out rather than worked out.
package policy

// Role is what someone holds on a project.
type Role string

const (
	RoleOwner       Role = "owner"
	RoleImplementer Role = "implementer"
	RoleEditor      Role = "editor"
	RoleViewer      Role = "viewer"
)

// Roles is every role, in the order the designs list them.
var Roles = []Role{RoleOwner, RoleImplementer, RoleEditor, RoleViewer}

// OrganisationRole is what someone holds on an organisation. Owning one
// carries the owner role on every project in it.
type OrganisationRole string

const (
	OrganisationOwner  OrganisationRole = "owner"
	OrganisationMember OrganisationRole = "member"
)

// Action is something someone might want to do.
type Action string

const (
	ActionReadContent        Action = "read-content"
	ActionEditContent        Action = "edit-content"
	ActionEditMedia          Action = "edit-media"
	ActionPublish            Action = "publish"
	ActionGoBackToAVersion   Action = "go-back-to-a-version"
	ActionManagePeople       Action = "manage-people"
	ActionChangeSettings     Action = "change-settings"
	ActionManageWorkflows    Action = "manage-workflows"
	ActionSeeTechnicalDetail Action = "see-technical-detail"
	ActionManageBilling      Action = "manage-billing"
	ActionDeleteProject      Action = "delete-project"
	ActionCreateProject      Action = "create-project"

	// These are held over the installation rather than over anything in it.
	// They are answered from whether the actor administers it, and the
	// resource says nothing about them.
	ActionCreateOrganisation       Action = "create-organisation"
	ActionRenameOrganisation       Action = "rename-organisation"
	ActionRemoveOrganisation       Action = "remove-organisation"
	ActionSeeEveryOrganisation     Action = "see-every-organisation"
	ActionManageOrganisationPeople Action = "manage-organisation-people"
	ActionAdministerInstallation   Action = "administer-installation"
)

// overTheInstallation is every action held over the installation itself. An
// administrator may do these and nothing else follows from it: administering
// is not editing, so none of these appears in the table of roles below.
var overTheInstallation = map[Action]bool{
	ActionCreateOrganisation:       true,
	ActionRenameOrganisation:       true,
	ActionRemoveOrganisation:       true,
	ActionSeeEveryOrganisation:     true,
	ActionManageOrganisationPeople: true,
	ActionAdministerInstallation:   true,
}

// OverTheInstallation says whether an action is held over the installation
// rather than over an organisation or a project.
func OverTheInstallation(action Action) bool {
	return overTheInstallation[action]
}

// Actions is every action, which the tests walk.
var Actions = []Action{
	ActionReadContent,
	ActionEditContent,
	ActionEditMedia,
	ActionPublish,
	ActionGoBackToAVersion,
	ActionManagePeople,
	ActionChangeSettings,
	ActionManageWorkflows,
	ActionSeeTechnicalDetail,
	ActionManageBilling,
	ActionDeleteProject,
	ActionCreateProject,
	ActionCreateOrganisation,
	ActionRenameOrganisation,
	ActionRemoveOrganisation,
	ActionSeeEveryOrganisation,
	ActionManageOrganisationPeople,
	ActionAdministerInstallation,
}

// allowed is the table from docs/design/DESIGN.md section 6. A role allows
// exactly the actions listed against it.
var allowed = map[Role]map[Action]bool{
	RoleOwner: {
		ActionReadContent:        true,
		ActionEditContent:        true,
		ActionEditMedia:          true,
		ActionPublish:            true,
		ActionGoBackToAVersion:   true,
		ActionManagePeople:       true,
		ActionChangeSettings:     true,
		ActionManageWorkflows:    true,
		ActionSeeTechnicalDetail: true,
		ActionManageBilling:      true,
		ActionDeleteProject:      true,
		ActionCreateProject:      true,
	},
	RoleImplementer: {
		ActionReadContent:        true,
		ActionEditContent:        true,
		ActionEditMedia:          true,
		ActionPublish:            true,
		ActionGoBackToAVersion:   true,
		ActionManagePeople:       true,
		ActionChangeSettings:     true,
		ActionManageWorkflows:    true,
		ActionSeeTechnicalDetail: true,
	},
	RoleEditor: {
		ActionReadContent:      true,
		ActionEditContent:      true,
		ActionEditMedia:        true,
		ActionPublish:          true,
		ActionGoBackToAVersion: true,
	},
	RoleViewer: {
		ActionReadContent: true,
	},
}

// Membership is one role someone holds.
type Membership struct {
	OrganisationID string
	ProjectID      string
	Role           Role
}

// Actor is the person asking, with everything needed to answer. An actor with
// no identifier is nobody, and nobody may do nothing.
type Actor struct {
	PersonID string
	// Administers says this person administers the installation. It is a
	// property of the account rather than a role on anything, because it is
	// held over the whole installation and over no part of it.
	Administers bool
	// Organisations maps an organisation's identifier to the role held there.
	Organisations map[string]OrganisationRole
	// Projects maps a project's identifier to the role held there.
	Projects map[string]Role
}

// Nobody is an actor who is not signed in.
func Nobody() Actor {
	return Actor{}
}

// Resource is what is being acted on. It names the organisation and the
// project, and may one day name something inside the project; the field
// exists now so that adding per-document permissions is a change to this
// package rather than to every caller.
type Resource struct {
	OrganisationID string
	ProjectID      string
	// Within names something inside the project, such as a page. Milestone 1
	// does not read it.
	Within string
}

// Project names a whole project.
func Project(organisationID, projectID string) Resource {
	return Resource{OrganisationID: organisationID, ProjectID: projectID}
}

// Organisation names a whole organisation.
func Organisation(organisationID string) Resource {
	return Resource{OrganisationID: organisationID}
}

// Can is the only place a role is compared.
func Can(actor Actor, action Action, resource Resource) bool {
	if actor.PersonID == "" {
		return false
	}

	// An action held over the installation is answered from that alone, and
	// administering grants nothing on a resource: whoever runs the service
	// manages who may reach a customer's site without thereby being able to
	// read or change it.
	if overTheInstallation[action] {
		return actor.Administers
	}

	role, held := EffectiveRole(actor, resource)
	if !held {
		return false
	}
	return allowed[role][action]
}

// EffectiveRole is the role an actor holds on a resource: the role they hold
// on the project, raised to owner when they own the organisation it is in.
func EffectiveRole(actor Actor, resource Resource) (Role, bool) {
	if actor.PersonID == "" {
		return "", false
	}

	if resource.OrganisationID != "" {
		if actor.Organisations[resource.OrganisationID] == OrganisationOwner {
			return RoleOwner, true
		}
	}

	if resource.ProjectID == "" {
		return "", false
	}

	role, held := actor.Projects[resource.ProjectID]
	if !held {
		return "", false
	}
	return role, true
}

// Allowed lists what an actor may do with a resource, which is what the
// interface is told so that it leaves out what it must rather than showing it
// disabled.
func Allowed(actor Actor, resource Resource) []Action {
	var out []Action
	for _, action := range Actions {
		// What somebody may do over the installation is not something they
		// may do to this resource, and a screen asking what it may offer on
		// a project should not be told about either.
		if overTheInstallation[action] {
			continue
		}
		if Can(actor, action, resource) {
			out = append(out, action)
		}
	}
	return out
}
