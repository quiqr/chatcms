import { maybe, type Key } from "../lib/say";

// The roles a person may hold on a project, as the server names them. They
// are compared as text here because the server is where a role is decided;
// this only chooses what to draw. Somebody who administers the organisation
// reaches a project as its implementer, so they are covered by the same set.
const seesSetup = new Set<string>(["implementer"]);
const seesTechnicalDetail = seesSetup;

export function canSeeSetup(role: string) {
  return seesSetup.has(role);
}

export function canSeeTechnicalDetail(role: string) {
  return seesTechnicalDetail.has(role);
}

/**
 * named is what a role is called on a screen. The server names roles in its
 * own words, which are not words anybody has to read.
 */
export function named(role: string): string {
  const key = `role.${role}` as Key;
  return maybe(key) ?? role;
}
