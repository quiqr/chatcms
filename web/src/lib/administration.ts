import { get, post, remove } from "./api";

export type AdministeredOrganisation = {
  slug: string;
  name: string;
  projects: number;
  people: number;
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

export function listEveryOrganisation() {
  return get<{ organisations: AdministeredOrganisation[] }>("/api/administration/organisations");
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
  return get<{ administers: boolean }>("/api/administration/me");
}

export function turnAccountOff(email: string) {
  return post<void>(`/api/administration/accounts/${encodeURIComponent(email)}/off`);
}

export function turnAccountOn(email: string) {
  return post<void>(`/api/administration/accounts/${encodeURIComponent(email)}/on`);
}
