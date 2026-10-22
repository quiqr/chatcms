import type { Role } from "../demo/content";

const seesSetup = new Set<Role>(["owner", "implementer"]);
const seesTechnicalDetail = seesSetup;

export function canSeeSetup(role: Role) {
  return seesSetup.has(role);
}

export function canSeeTechnicalDetail(role: Role) {
  return seesTechnicalDetail.has(role);
}
