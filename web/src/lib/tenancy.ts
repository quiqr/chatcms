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

export type WordPart = "kept" | "gone" | "added";

export type Word = {
  part: WordPart;
  text: string;
};

export type FieldChange = {
  name: string;
  words?: Word[];
};

export type ChangeKind =
  | "changed-page"
  | "new-page"
  | "removed-page"
  | "new-image"
  | "removed-image"
  | "other-file";

export type Change = {
  kind: ChangeKind;
  section: string;
  page: string;
  by?: string;
  at?: string;
  fields?: FieldChange[];
};

export type Review = {
  count: number;
  lastPublished?: string;
  changes: Change[];
};

export type PublishStep = {
  title: string;
  state: "succeeded" | "failed" | "not-run";
};

export type Publish = {
  id: string;
  state: "running" | "succeeded" | "failed" | "interrupted";
  note?: string;
  changeCount: number;
  startedAt: string;
  finishedAt?: string;
  steps?: PublishStep[];
  trouble?: string;
  detail?: string;
};

export function readReview(organisation: string, name: string) {
  return get<Review>(`${project(organisation, name)}/review`);
}

export function publishNow(organisation: string, name: string, note: string) {
  return post<Publish>(`${project(organisation, name)}/publishes`, { note });
}

export function readPublishes(organisation: string, name: string) {
  return get<{ publishes: Publish[] }>(`${project(organisation, name)}/publishes`);
}

export type Version = {
  version: string;
  title: string;
  by?: string;
  at?: string;
  changes: number;
  live: boolean;
};

export type VersionGroup = {
  period: string;
  entries: Version[];
};

export type VersionTimeline = {
  unpublished: number;
  groups: VersionGroup[];
};

export type GoBackWarning = {
  title: string;
  when?: string;
  setAside: number;
  reassurance: string;
};

export type WentBack = {
  title: string;
  published?: Publish;
};

export function readTimeline(organisation: string, name: string) {
  return get<VersionTimeline>(`${project(organisation, name)}/versions`);
}

export function readVersion(organisation: string, name: string, version: string) {
  return get<Review>(`${project(organisation, name)}/versions/${version}`);
}

export function beforeGoingBack(organisation: string, name: string, version: string) {
  return get<GoBackWarning>(`${project(organisation, name)}/versions/${version}/go-back`);
}

export function goBackTo(
  organisation: string,
  name: string,
  version: string,
  publishRightAway: boolean,
) {
  return post<WentBack>(`${project(organisation, name)}/versions/${version}/go-back`, {
    publishRightAway,
  });
}
