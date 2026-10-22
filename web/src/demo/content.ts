export const PLACEHOLDER = true;

export type Role = "owner" | "implementer" | "editor" | "viewer";

export type DemoProject = {
  slug: string;
  name: string;
  domain: string;
  unpublished: number;
  people: number;
  lastActivity: string;
};

export type DemoPerson = {
  name: string;
  email: string;
  role: Role;
  lastActive: string;
  pending?: boolean;
};

export const demoOrganisation = {
  slug: "technative",
  name: "TechNative",
};

export const demoProjects: DemoProject[] = [
  {
    slug: "korenaar",
    name: "Korenaar Bakery",
    domain: "korenaar.example",
    unpublished: 4,
    people: 3,
    lastActivity: "12 min ago",
  },
  {
    slug: "studio-noord",
    name: "Studio Noord",
    domain: "studionoord.example",
    unpublished: 0,
    people: 2,
    lastActivity: "Tuesday at 9:12",
  },
];

export const demoCollections = [
  { key: "articles", title: "News", itemName: "article", count: 12 },
  { key: "breads", title: "Breads", itemName: "bread", count: 7 },
];

export const demoPages = [
  { key: "home", title: "Homepage" },
  { key: "about", title: "About us" },
];

export const demoPeople: DemoPerson[] = [
  { name: "Joris Bakker", email: "joris@korenaar.example", role: "owner", lastActive: "12 min ago" },
  { name: "Marieke de Vries", email: "marieke@korenaar.example", role: "editor", lastActive: "Tuesday at 9:12" },
  { name: "Sam Visser", email: "sam@korenaar.example", role: "viewer", lastActive: "Not yet", pending: true },
];

export const demoViewer = {
  name: "Joris Bakker",
  role: "owner" as Role,
};
