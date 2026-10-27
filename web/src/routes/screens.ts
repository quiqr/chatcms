// chatcms-not-a-string: the name the design hand-off gives each screen.
// Only the route test reads these; no screen shows one.
export type ScreenName =
  | "sign-in"
  | "projects"
  | "onboarding"
  | "people"
  | "content"
  | "editor"
  | "media"
  | "publish"
  | "history"
  | "sync";

export type Screen = {
  name: ScreenName;
  title: string;
  path: string;
};

export const screens: Screen[] = [
  { name: "sign-in", title: "Sign in", path: "/sign-in" },
  { name: "projects", title: "Projects", path: "/projects" },
  { name: "onboarding", title: "Set up a project", path: "/o/:organisation/p/:project/setup" },
  { name: "people", title: "People & roles", path: "/o/:organisation/p/:project/people" },
  { name: "content", title: "Content", path: "/o/:organisation/p/:project/content" },
  { name: "editor", title: "Edit a page", path: "/o/:organisation/p/:project/content/:page" },
  { name: "media", title: "Media", path: "/o/:organisation/p/:project/media" },
  { name: "publish", title: "Publish", path: "/o/:organisation/p/:project/publish" },
  { name: "history", title: "History", path: "/o/:organisation/p/:project/history" },
  { name: "sync", title: "Sync status", path: "/o/:organisation/p/:project/sync" },
];

export function pathFor(
  name: ScreenName,
  organisation: string,
  project = "korenaar",
  page = "about",
) {
  const screen = screens.find((candidate) => candidate.name === name);
  if (!screen) throw new Error(`There is no screen called ${name}`);
  return screen.path
    .replace(":organisation", organisation)
    .replace(":project", project)
    .replace(":page", page);
}
