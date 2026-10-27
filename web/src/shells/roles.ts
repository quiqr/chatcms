// The roles a person may hold on a project, as the server names them. They
// are compared as text here because the server is where a role is decided;
// this only chooses what to draw.
const seesSetup = new Set<string>(["owner", "implementer"]);
const seesTechnicalDetail = seesSetup;

export function canSeeSetup(role: string) {
  return seesSetup.has(role);
}

export function canSeeTechnicalDetail(role: string) {
  return seesTechnicalDetail.has(role);
}
