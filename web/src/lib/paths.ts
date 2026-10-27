/**
 * Every address in the application is built here. A project's address is only
 * unique within its organisation, so the organisation is part of it; building
 * these by hand in each screen is how one of them quietly drops it.
 */

export function projectsPath(organisation: string) {
  return `/o/${organisation}/projects`;
}

export function projectPath(organisation: string, project: string, screen = "content") {
  return `/o/${organisation}/p/${project}/${screen}`;
}

export function pagePath(organisation: string, project: string, page: string) {
  return `/o/${organisation}/p/${project}/content/${page}`;
}

export function organisationPath(organisation: string) {
  return `/o/${organisation}/settings`;
}
