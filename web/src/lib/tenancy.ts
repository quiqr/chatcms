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
