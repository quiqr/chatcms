import { get, post, put, remove } from "./api";
import type { Usage } from "../components/ui/usage";

export type AdministeredOrganisation = {
  slug: string;
  name: string;
  projects: number;
  people: number;
  usage: Usage;
  admins: { email: string; displayName: string }[];
  madeBy?: string;
  madeOn?: string;
};

export type Administrator = {
  email: string;
  displayName: string;
  yourself: boolean;
};

export type AdministrationRecord = {
  who: string;
  did: string;
  to: string;
  detail?: Record<string, string>;
  at: string;
};

export function listEveryOrganisation(search = "", filter = "") {
  const asked = new URLSearchParams();
  if (search.trim() !== "") asked.set("search", search.trim());
  if (filter !== "") asked.set("filter", filter);
  const question = asked.toString();

  return get<{ organisations: AdministeredOrganisation[] }>(
    `/api/administration/organisations${question === "" ? "" : `?${question}`}`,
  );
}

export function makeOrganisation(body: { slug: string; name: string; ownerEmail: string }) {
  return post<{ slug: string; name: string }>("/api/administration/organisations", body);
}

export function putSomebodyInto(organisation: string, email: string, role: string) {
  return post<void>(`/api/administration/organisations/${organisation}/people`, { email, role });
}

export function takeSomebodyOut(organisation: string, email: string) {
  return remove<void>(
    `/api/administration/organisations/${organisation}/people/${encodeURIComponent(email)}`,
  );
}

export function listAdministrators() {
  return get<{ administrators: Administrator[] }>("/api/administration/administrators");
}

export function appoint(email: string) {
  return post<void>("/api/administration/administrators", { email });
}

export function standDown(email: string) {
  return remove<void>(`/api/administration/administrators/${encodeURIComponent(email)}`);
}

export function listRecords() {
  return get<{ records: AdministrationRecord[] }>("/api/administration/records");
}

export function mayAdminister() {
  return get<{ administers: boolean; organisations?: number; users?: number }>(
    "/api/administration/me",
  );
}

export function turnAccountOff(email: string) {
  return post<void>(`/api/administration/accounts/${encodeURIComponent(email)}/off`);
}

export function turnAccountOn(email: string) {
  return post<void>(`/api/administration/accounts/${encodeURIComponent(email)}/on`);
}

export type OrganisationWhole = {
  slug: string;
  name: string;
  usage: Usage;
  people: { personId: string; email: string; displayName: string; role: string; stillToAccept: boolean }[];
  projects: { slug: string; name: string; domain: string }[];
};

export function readOrganisation(organisation: string) {
  return get<OrganisationWhole>(`/api/administration/organisations/${organisation}`);
}

export function removeOrganisation(organisation: string) {
  return remove<void>(`/api/administration/organisations/${organisation}`);
}

export type Belonging = {
  slug: string;
  name: string;
  organisation?: string;
  role: string;
};

export type PersonWhole = {
  personId: string;
  email: string;
  displayName: string;
  administers: boolean;
  accountOn: boolean;
  stillToAccept: boolean;
  lastSeen?: string;
  organisations: Belonging[];
  projects: Belonging[];
};

export function listEveryPerson(search: string, only = "", after = "", most = 0) {
  const asked = new URLSearchParams();
  if (search.trim() !== "") asked.set("search", search.trim());
  if (only !== "") asked.set("only", only);
  if (after !== "") asked.set("after", after);
  if (most > 0) asked.set("most", String(most));
  const question = asked.toString();

  return get<{ people: PersonWhole[]; shown: number; altogether: number; after?: string }>(
    `/api/administration/people${question === "" ? "" : `?${question}`}`,
  );
}

export function changeAnAccount(email: string, displayName: string, address: string) {
  return put<void>(`/api/administration/people/${encodeURIComponent(email)}`, {
    displayName,
    email: address,
  });
}

export function sendAPasswordReset(email: string) {
  return post<void>(`/api/administration/people/${encodeURIComponent(email)}/password-reset`);
}

export function deleteAnAccount(email: string) {
  return remove<void>(`/api/administration/people/${encodeURIComponent(email)}`);
}

export function readPerson(email: string) {
  return get<PersonWhole>(`/api/administration/people/${encodeURIComponent(email)}`);
}

export function makePerson(body: { email: string; displayName: string }) {
  return post<PersonWhole>("/api/administration/people", body);
}

export function actAs(email: string) {
  return post<void>(`/api/administration/people/${encodeURIComponent(email)}/act-as`);
}

export function setLimits(organisation: string, members: number | null, projects: number | null) {
  return put<void>(`/api/administration/organisations/${organisation}/limits`, {
    members,
    projects,
  });
}
