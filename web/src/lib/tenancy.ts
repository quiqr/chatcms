import { get, post, put, remove } from "./api";

export type Organisation = {
  slug: string;
  name: string;
};

export type Project = {
  slug: string;
  name: string;
  domain: string;
  yourRole: string;
  youMay: string[];
  unpublished: number;
};

export type Member = {
  personId: string;
  email: string;
  displayName: string;
  role: string;
  stillToAccept: boolean;
};

export const roles = ["owner", "implementer", "editor", "viewer"] as const;
export type RoleName = (typeof roles)[number];

export function listOrganisations() {
  return get<{ organisations: Organisation[] }>("/api/organisations");
}

export function listProjects(organisation: string) {
  return get<{ projects: Project[] }>(`/api/organisations/${organisation}/projects`);
}

export function listPeople(organisation: string, project: string) {
  return get<{ people: Member[] }>(
    `/api/organisations/${organisation}/projects/${project}/people`,
  );
}

export function invitePerson(organisation: string, project: string, email: string, role: string) {
  return post<Member>(`/api/organisations/${organisation}/projects/${project}/people`, {
    email,
    role,
  });
}

export function changeRole(
  organisation: string,
  project: string,
  personId: string,
  role: string,
) {
  return put<Member>(
    `/api/organisations/${organisation}/projects/${project}/people/${personId}`,
    { role },
  );
}

export function removePerson(organisation: string, project: string, personId: string) {
  return remove<void>(
    `/api/organisations/${organisation}/projects/${project}/people/${personId}`,
  );
}

export type DeployKey = {
  publicKey: string;
  advice: string;
};

export type SiteReport = {
  kind: string;
  remote?: string;
  isHugoSite: boolean;
  hasModel: boolean;
  ready: boolean;
  pages: number;
  collections: number;
  problem?: string;
  lookedFor?: string[];
};

export type Connection = {
  good: boolean;
  message: string;
  publicKey?: string;
};

export type Publishing = {
  to: "main-line" | "its-own-line";
  branch?: string;
  form: "source" | "built";
};

export const defaultPublishBranch = "quiqr/publish";

function project(organisation: string, name: string) {
  return `/api/organisations/${organisation}/projects/${name}`;
}

export function readKey(organisation: string, name: string) {
  return get<DeployKey>(`${project(organisation, name)}/key`);
}

export function makeKey(organisation: string, name: string) {
  return post<DeployKey>(`${project(organisation, name)}/key`);
}

export function readSite(organisation: string, name: string) {
  return get<SiteReport>(`${project(organisation, name)}/site`);
}

export function importSite(
  organisation: string,
  name: string,
  body: { kind: string; address: string; remote?: string },
) {
  return post<SiteReport>(`${project(organisation, name)}/import`, body);
}

export function testConnection(organisation: string, name: string) {
  return post<Connection>(`${project(organisation, name)}/connection`);
}

export function readPublishing(organisation: string, name: string) {
  return get<Publishing>(`${project(organisation, name)}/publishing`);
}

export function setPublishing(organisation: string, name: string, body: Publishing) {
  return put<Publishing>(`${project(organisation, name)}/publishing`, body);
}
