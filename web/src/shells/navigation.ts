import {
  FileText,
  History,
  Image,
  SlidersHorizontal,
  Upload,
  Users,
  Workflow,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type NavEntry = {
  key: string;
  label: string;
  to: string;
  icon: LucideIcon;
  soon?: boolean;
};

export function siteNavigation(organisation: string, projectSlug: string): NavEntry[] {
  return [
    { key: "content", label: "Content", to: `/o/${organisation}/p/${projectSlug}/content`, icon: FileText },
    { key: "media", label: "Media", to: `/o/${organisation}/p/${projectSlug}/media`, icon: Image },
    { key: "publish", label: "Publish", to: `/o/${organisation}/p/${projectSlug}/publish`, icon: Upload },
    { key: "history", label: "History", to: `/o/${organisation}/p/${projectSlug}/history`, icon: History },
  ];
}

export function setupNavigation(organisation: string, projectSlug: string): NavEntry[] {
  return [
    { key: "people", label: "People & roles", to: `/o/${organisation}/p/${projectSlug}/people`, icon: Users },
    { key: "settings", label: "Project settings", to: `/o/${organisation}/p/${projectSlug}/settings`, icon: SlidersHorizontal },
    { key: "workflows", label: "Workflows", to: `/o/${organisation}/p/${projectSlug}/workflows`, icon: Workflow, soon: true },
  ];
}
