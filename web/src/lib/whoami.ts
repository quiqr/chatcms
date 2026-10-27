import { createContext, useContext } from "react";

import { get, post, put, remove } from "./api";

export type Person = {
  id: string;
  email: string;
  displayName: string;
};

export type WhoAmI = {
  person: Person;
  administers: boolean;
  organisations: { slug: string; name: string }[];
  /** appearance is what they chose, or empty to follow the machine. */
  appearance?: string;
};

export type OrganisationMember = {
  personId: string;
  email: string;
  displayName: string;
  role: string;
  stillToAccept: boolean;
  lastSeen?: string;
};

export function whoAmI() {
  return get<WhoAmI>("/api/whoami");
}

export function renameOrganisation(organisation: string, name: string) {
  return put<{ slug: string; name: string }>(`/api/organisations/${organisation}`, { name });
}

export function listOrganisationPeople(organisation: string) {
  return get<{ people: OrganisationMember[] }>(`/api/organisations/${organisation}/people`);
}

export type Here = {
  who: WhoAmI;
  /** organisation is the one the address names. */
  organisation: string;
  /** again reloads it, after a rename or a new organisation. */
  again: () => void;
};

/**
 * HereContext carries what every screen needs before it can draw its shell:
 * who is looking, what they may switch to, and which organisation the
 * address names. It is asked once and shared, because three requests per
 * navigation would be three chances to half-load.
 */
export const HereContext = createContext<Here | null>(null);

export function useHere() {
  const here = useContext(HereContext);
  if (!here) {
    throw new Error("useHere was called outside the provider that loads it");
  }
  return here;
}

export function inviteToOrganisation(organisation: string, email: string, role: string) {
  return post<OrganisationMember>(`/api/organisations/${organisation}/people`, { email, role });
}

export function changeOrganisationRole(organisation: string, personId: string, role: string) {
  return put<void>(`/api/organisations/${organisation}/people/${personId}`, { role });
}

export function removeFromOrganisation(organisation: string, personId: string) {
  return remove<void>(`/api/organisations/${organisation}/people/${personId}`);
}

export function sendTheInvitationAgain(organisation: string, personId: string) {
  return post<void>(`/api/organisations/${organisation}/people/${personId}/invitation`);
}
