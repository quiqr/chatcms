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
  { name: "onboarding", title: "Set up a project", path: "/p/:project/setup" },
  { name: "people", title: "People & roles", path: "/p/:project/people" },
  { name: "content", title: "Content", path: "/p/:project/content" },
  { name: "editor", title: "Edit a page", path: "/p/:project/content/:page" },
  { name: "media", title: "Media", path: "/p/:project/media" },
  { name: "publish", title: "Publish", path: "/p/:project/publish" },
  { name: "history", title: "History", path: "/p/:project/history" },
  { name: "sync", title: "Sync status", path: "/p/:project/sync" },
];

export function pathFor(name: ScreenName, project = "korenaar", page = "about") {
  const screen = screens.find((candidate) => candidate.name === name);
  if (!screen) throw new Error(`There is no screen called ${name}`);
  return screen.path.replace(":project", project).replace(":page", page);
}
